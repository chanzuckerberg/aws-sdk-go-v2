package manager

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// MaxUploadParts is the maximum allowed number of parts in a multi-part upload
// on Amazon S3.
const MaxUploadParts int32 = 10000

// MinUploadPartSize is the minimum allowed part size when uploading a part to
// Amazon S3.
const MinUploadPartSize int64 = 1024 * 1024 * 5

// DefaultUploadPartSize is the default part size to buffer chunks of a
// payload into.
const DefaultUploadPartSize = MinUploadPartSize

// DefaultUploadConcurrency is the default number of goroutines to spin up when
// using Upload().
const DefaultUploadConcurrency = 5

// A MultiUploadFailure wraps a failed S3 multipart upload. An error returned
// will satisfy this interface when a multi part upload failed to upload all
// chucks to S3. In the case of a failure the UploadID is needed to operate on
// the chunks, if any, which were uploaded.
//
// Example:
//
//	u := manager.NewUploader(client)
//	output, err := u.upload(context.Background(), input)
//	if err != nil {
//		var multierr manager.MultiUploadFailure
//		if errors.As(err, &multierr) {
//			fmt.Printf("upload failure UploadID=%s, %s\n", multierr.UploadID(), multierr.Error())
//		} else {
//			fmt.Printf("upload failure, %s\n", err.Error())
//		}
//	}
//
type MultiUploadFailure interface {
	error

	// UploadID returns the upload id for the S3 multipart upload that failed.
	UploadID() string
}

// A multiUploadError wraps the upload ID of a failed s3 multipart upload.
// Composed of BaseError for code, message, and original error
//
// Should be used for an error that occurred failing a S3 multipart upload,
// and a upload ID is available. If an uploadID is not available a more relevant
type multiUploadError struct {
	err error

	// ID for multipart upload which failed.
	uploadID string
}

// batchItemError returns the string representation of the error.
//
// See apierr.BaseError ErrorWithExtra for output format
//
// Satisfies the error interface.
func (m *multiUploadError) Error() string {
	var extra string
	if m.err != nil {
		extra = fmt.Sprintf(", cause: %s", m.err.Error())
	}
	return fmt.Sprintf("upload multipart failed, upload id: %s%s", m.uploadID, extra)
}

// Unwrap returns the underlying error that cause the upload failure
func (m *multiUploadError) Unwrap() error {
	return m.err
}

// UploadID returns the id of the S3 upload which failed.
func (m *multiUploadError) UploadID() string {
	return m.uploadID
}

// UploadOutput represents a response from the Upload() call.
type UploadOutput struct {
	// The URL where the object was uploaded to.
	Location string

	// The version of the object that was uploaded. Will only be populated if
	// the S3 Bucket is versioned. If the bucket is not versioned this field
	// will not be set.
	VersionID *string

	// The ID for a multipart upload to S3. In the case of an error the error
	// can be cast to the MultiUploadFailure interface to extract the upload ID.
	UploadID string
}

// WithUploaderRequestOptions appends to the Uploader's API client options.
func WithUploaderRequestOptions(opts ...func(*s3.Options)) func(*Uploader) {
	return func(u *Uploader) {
		u.ClientOptions = append(u.ClientOptions, opts...)
	}
}

// The Uploader structure that calls Upload(). It is safe to call Upload()
// on this structure for multiple objects and across concurrent goroutines.
// Mutating the Uploader's properties is not safe to be done concurrently.
type Uploader struct {
	// The buffer size (in bytes) to use when buffering data into chunks and
	// sending them as parts to S3. The minimum allowed part size is 5MB, and
	// if this value is set to zero, the DefaultUploadPartSize value will be used.
	PartSize int64

	// The number of goroutines to spin up in parallel per call to Upload when
	// sending parts. If this is set to zero, the DefaultUploadConcurrency value
	// will be used.
	//
	// The concurrency pool is not shared between calls to Upload.
	Concurrency int

	// Setting this value to true will cause the SDK to avoid calling
	// AbortMultipartUpload on a failure, leaving all successfully uploaded
	// parts on S3 for manual recovery.
	//
	// Note that storing parts of an incomplete multipart upload counts towards
	// space usage on S3 and will add additional costs if not cleaned up.
	LeavePartsOnError bool

	// MaxUploadParts is the max number of parts which will be uploaded to S3.
	// Will be used to calculate the partsize of the object to be uploaded.
	// E.g: 5GB file, with MaxUploadParts set to 100, will upload the file
	// as 100, 50MB parts. With a limited of s3.MaxUploadParts (10,000 parts).
	//
	// MaxUploadParts must not be used to limit the total number of bytes uploaded.
	// Use a type like to io.LimitReader (https://golang.org/pkg/io/#LimitedReader)
	// instead. An io.LimitReader is helpful when uploading an unbounded reader
	// to S3, and you know its maximum size. Otherwise the reader's io.EOF returned
	// error must be used to signal end of stream.
	//
	// Defaults to package const's MaxUploadParts value.
	MaxUploadParts int32

	// The client to use when uploading to S3.
	S3 UploadAPIClient

	// List of request options that will be passed down to individual API
	// operation requests made by the uploader.
	ClientOptions []func(*s3.Options)

	// Defines the buffer strategy used when uploading a part
	BufferProvider ReadSeekerWriteToProvider

	// partPool allows for the re-usage of streaming payload part buffers between upload calls
	partPool byteSlicePool
}

// NewUploader creates a new Uploader instance to upload objects to S3. Pass In
// additional functional options to customize the uploader's behavior. Requires a
// client.ConfigProvider in order to create a S3 service client. The session.Session
// satisfies the client.ConfigProvider interface.
//
// Example:
//	// Load AWS Config
//	cfg, err := config.LoadDefaultConfig(context.TODO())
//	if err != nil {
//		panic(err)
//	}
//
//	// Create an S3 Client with the config
//	client := s3.NewFromConfig(cfg)
//
//	// Create an uploader passing it the client
//  uploader := manager.NewUploader(client)
//
//	// Create an uploader with the client and custom options
//	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
//		u.PartSize = 64 * 1024 * 1024 // 64MB per part
//	})
func NewUploader(client UploadAPIClient, options ...func(*Uploader)) *Uploader {
	u := &Uploader{
		S3:                client,
		PartSize:          DefaultUploadPartSize,
		Concurrency:       DefaultUploadConcurrency,
		LeavePartsOnError: false,
		MaxUploadParts:    MaxUploadParts,
		BufferProvider:    defaultUploadBufferProvider(),
	}

	for _, option := range options {
		option(u)
	}

	u.partPool = newByteSlicePool(u.PartSize)

	return u
}

func (u Uploader) uploadWithSingleUploader(i uploader, opts ...func(*Uploader)) (*UploadOutput, error) {
	// Copy ClientOptions
	clientOptions := make([]func(*s3.Options), 0, len(i.cfg.ClientOptions)+1)
	clientOptions = append(clientOptions, func(o *s3.Options) {
		o.APIOptions = append(o.APIOptions, middleware.AddSDKAgentKey(middleware.FeatureMetadata, userAgentKey))
	})
	clientOptions = append(clientOptions, i.cfg.ClientOptions...)
	i.cfg.ClientOptions = clientOptions

	for _, opt := range opts {
		opt(&i.cfg)
	}

	return i.upload()
}

// Upload uploads an object to S3, intelligently buffering large
// files into smaller chunks and sending them in parallel across multiple
// goroutines. You can configure the buffer size and concurrency through the
// Uploader parameters.
//
// Additional functional options can be provided to configure the individual
// upload. These options are copies of the Uploader instance Upload is called from.
// Modifying the options will not impact the original Uploader instance.
//
// Use the WithUploaderRequestOptions helper function to pass in request
// options that will be applied to all API operations made with this uploader.
//
// It is safe to call this method concurrently across goroutines.
func (u Uploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*Uploader)) (*UploadOutput, error) {
	return u.uploadWithSingleUploader(uploader{in: input, cfg: u, ctx: ctx}, opts...)

}

// ResumeUpload resumes an existing multipart upload to S3, intelligently buffering
// large files into smaller chunks and sending them in parallel across multiple
// goroutines. You can configure the buffer size and concurrency through the
// Uploader parameters. The parts that are already uploaded have their md5
// checkums computed locally and compared with their uploaded ETag. If these do
// not match the upload fails. This is to ensure the integrity of the resumed
// multipart upload in case the data or part size differs from the original
// multipart upload.
//
// Additional functional options can be provided to configure the individual
// upload. These options are copies of the Uploader instance Upload is called from.
// Modifying the options will not impact the original Uploader instance.
//
// Use the WithUploaderRequestOptions helper function to pass in request
// options that will be applied to all API operations made with this uploader.
//
// It is safe to call this method concurrently across goroutines.
func (u Uploader) ResumeUpload(ctx context.Context, input *s3.PutObjectInput, uploadID *string, opts ...func(*Uploader)) (*UploadOutput, error) {
	return u.uploadWithSingleUploader(uploader{in: input, cfg: u, existingUploadID: uploadID, ctx: ctx}, opts...)
}

// internal structure to manage an upload to S3.
type uploader struct {
	ctx context.Context
	cfg Uploader

	in               *s3.PutObjectInput
	existingUploadID *string

	readerPos int64 // current reader position
	totalSize int64 // set to -1 if the size is not known
}

// internal logic for deciding whether to upload a single part or use a
// multipart upload.
func (u *uploader) upload() (*UploadOutput, error) {
	if err := u.init(); err != nil {
		return nil, fmt.Errorf("unable to initialize upload: %w", err)
	}
	defer u.cfg.partPool.Close()

	if u.cfg.PartSize < MinUploadPartSize {
		return nil, fmt.Errorf("part size must be at least %d bytes", MinUploadPartSize)
	}

	// Do one read to determine if we have more than one part
	reader, _, cleanup, err := u.nextReader()
	if err == io.EOF { // single part
		return u.singlePart(reader, cleanup)
	} else if err != nil {
		cleanup()
		return nil, fmt.Errorf("read upload data failed: %w", err)
	}

	mu := multiuploader{uploader: u, eTagByPartNumber: make(map[int32]string)}
	return mu.upload(reader, cleanup)
}

// init will initialize all default options.
func (u *uploader) init() error {
	if u.cfg.Concurrency == 0 {
		u.cfg.Concurrency = DefaultUploadConcurrency
	}
	if u.cfg.PartSize == 0 {
		u.cfg.PartSize = DefaultUploadPartSize
	}
	if u.cfg.MaxUploadParts == 0 {
		u.cfg.MaxUploadParts = MaxUploadParts
	}

	// Try to get the total size for some optimizations
	if err := u.initSize(); err != nil {
		return err
	}

	// If PartSize was changed or partPool was never setup then we need to allocated a new pool
	// so that we return []byte slices of the correct size
	poolCap := u.cfg.Concurrency + 1
	if u.cfg.partPool == nil || u.cfg.partPool.SliceSize() != u.cfg.PartSize {
		u.cfg.partPool = newByteSlicePool(u.cfg.PartSize)
		u.cfg.partPool.ModifyCapacity(poolCap)
	} else {
		u.cfg.partPool = &returnCapacityPoolCloser{byteSlicePool: u.cfg.partPool}
		u.cfg.partPool.ModifyCapacity(poolCap)
	}

	return nil
}

// initSize tries to detect the total stream size, setting u.totalSize. If
// the size is not known, totalSize is set to -1.
func (u *uploader) initSize() error {
	u.totalSize = -1

	switch r := u.in.Body.(type) {
	case io.Seeker:
		n, err := seekerLen(r)
		if err != nil {
			return err
		}
		u.totalSize = n

		// Try to adjust partSize if it is too small and account for
		// integer division truncation.
		if u.totalSize/u.cfg.PartSize >= int64(u.cfg.MaxUploadParts) {
			// Add one to the part size to account for remainders
			// during the size calculation. e.g odd number of bytes.
			u.cfg.PartSize = (u.totalSize / int64(u.cfg.MaxUploadParts)) + 1
		}
	}

	return nil
}

// nextReader returns a seekable reader representing the next packet of data.
// This operation increases the shared u.readerPos counter, but note that it
// does not need to be wrapped in a mutex because nextReader is only called
// from the main thread.
func (u *uploader) nextReader() (io.ReadSeeker, int, func(), error) {
	switch r := u.in.Body.(type) {
	case readerAtSeeker:
		var err error

		n := u.cfg.PartSize
		if u.totalSize >= 0 {
			bytesLeft := u.totalSize - u.readerPos

			if bytesLeft <= u.cfg.PartSize {
				err = io.EOF
				n = bytesLeft
			}
		}

		var (
			reader  io.ReadSeeker
			cleanup func()
		)

		reader = io.NewSectionReader(r, u.readerPos, n)
		if u.cfg.BufferProvider != nil {
			reader, cleanup = u.cfg.BufferProvider.GetWriteTo(reader)
		} else {
			cleanup = func() {}
		}

		u.readerPos += n

		return reader, int(n), cleanup, err

	default:
		part, err := u.cfg.partPool.Get(u.ctx)
		if err != nil {
			return nil, 0, func() {}, err
		}

		n, err := readFillBuf(r, *part)
		u.readerPos += int64(n)

		cleanup := func() {
			u.cfg.partPool.Put(part)
		}

		return bytes.NewReader((*part)[0:n]), n, cleanup, err
	}
}

func readFillBuf(r io.Reader, b []byte) (offset int, err error) {
	for offset < len(b) && err == nil {
		var n int
		n, err = r.Read(b[offset:])
		offset += n
	}

	return offset, err
}

// singlePart contains upload logic for uploading a single chunk via
// a regular PutObject request. Multipart requests require at least two
// parts, or at least 5MB of data.
func (u *uploader) singlePart(r io.ReadSeeker, cleanup func()) (*UploadOutput, error) {
	defer cleanup()

	params := &s3.PutObjectInput{}
	awsutil.Copy(params, u.in)
	params.Body = r

	// Need to use request form because URL generated in request is
	// used in return.

	var locationRecorder recordLocationClient
	out, err := u.cfg.S3.PutObject(u.ctx, params, append(u.cfg.ClientOptions, locationRecorder.WrapClient())...)
	if err != nil {
		return nil, err
	}

	return &UploadOutput{
		Location:  locationRecorder.location,
		VersionID: out.VersionId,
	}, nil
}

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type recordLocationClient struct {
	httpClient
	location string
}

func (c *recordLocationClient) WrapClient() func(o *s3.Options) {
	return func(o *s3.Options) {
		c.httpClient = o.HTTPClient
		o.HTTPClient = c
	}
}

func (c *recordLocationClient) Do(r *http.Request) (resp *http.Response, err error) {
	resp, err = c.httpClient.Do(r)
	if err != nil {
		return resp, err
	}

	if resp.Request != nil && resp.Request.URL != nil {
		url := *resp.Request.URL
		url.RawQuery = ""
		c.location = url.String()
	}

	return resp, err
}

// internal structure to manage a specific multipart upload to S3.
type multiuploader struct {
	*uploader
	wg               sync.WaitGroup
	m                sync.Mutex
	err              error
	uploadID         string
	parts            completedParts
	eTagByPartNumber map[int32]string
}

// keeps track of a single chunk of data being sent to S3.
type chunk struct {
	buf     io.ReadSeeker
	num     int32
	cleanup func()
}

// completedParts is a wrapper to make parts sortable by their part number,
// since S3 required this list to be sent in sorted order.
type completedParts []types.CompletedPart

func (a completedParts) Len() int           { return len(a) }
func (a completedParts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a completedParts) Less(i, j int) bool { return a[i].PartNumber < a[j].PartNumber }

// upload will perform a multipart upload using the firstBuf buffer containing
// the first chunk of data.
func (u *multiuploader) upload(firstBuf io.ReadSeeker, cleanup func()) (*UploadOutput, error) {
	var err error
	var locationRecorder recordLocationClient
	if u.uploader.existingUploadID != nil {
		u.uploadID = *u.uploader.existingUploadID
		params := &s3.ListPartsInput{}
		awsutil.Copy(params, u.in)
		params.UploadId = u.uploader.existingUploadID
		paginator := s3.NewListPartsPaginator(u.cfg.S3, params)
		repeat := false
		for paginator.HasMorePages() && !repeat {
			parts, err := paginator.NextPage(u.ctx, append(u.cfg.ClientOptions, locationRecorder.WrapClient())...)
			if err != nil {
				return nil, err
			}
			for _, part := range parts.Parts {
				// HACK: currently the paginator will loop, paginating forever
				// This logic stops the infinite loop when we see the same part again
				// This can be removed when https://github.com/aws/aws-sdk-go-v2/issues/1140 is resolved
				if _, repeat = u.eTagByPartNumber[part.PartNumber]; repeat {
					break
				}
				eTag, err := strconv.Unquote(*part.ETag)
				if err != nil {
					return nil, err
				}
				u.eTagByPartNumber[part.PartNumber] = eTag
			}
		}
	} else {
		params := &s3.CreateMultipartUploadInput{}
		awsutil.Copy(params, u.in)

		// Create the multipart
		resp, err := u.cfg.S3.CreateMultipartUpload(u.ctx, params, append(u.cfg.ClientOptions, locationRecorder.WrapClient())...)
		if err != nil {
			cleanup()
			return nil, err
		}
		u.uploadID = *resp.UploadId
	}

	// Create the workers
	ch := make(chan chunk, u.cfg.Concurrency)
	for i := 0; i < u.cfg.Concurrency; i++ {
		u.wg.Add(1)
		go u.readChunk(ch)
	}

	// Send part 1 to the workers
	var num int32 = 1
	ch <- chunk{buf: firstBuf, num: num, cleanup: cleanup}

	// Read and queue the rest of the parts
	for u.geterr() == nil && err == nil {
		var (
			reader       io.ReadSeeker
			nextChunkLen int
			ok           bool
		)

		reader, nextChunkLen, cleanup, err = u.nextReader()
		ok, err = u.shouldContinue(num, nextChunkLen, err)
		if !ok {
			cleanup()
			if err != nil {
				u.seterr(err)
			}
			break
		}

		num++

		ch <- chunk{buf: reader, num: num, cleanup: cleanup}
	}

	// Close the channel, wait for workers, and complete upload
	close(ch)
	u.wg.Wait()
	complete := u.complete()

	if err := u.geterr(); err != nil {
		return nil, &multiUploadError{
			err:      err,
			uploadID: u.uploadID,
		}
	}

	return &UploadOutput{
		Location:  locationRecorder.location,
		VersionID: complete.VersionId,
		UploadID:  u.uploadID,
	}, nil
}

func (u *multiuploader) shouldContinue(part int32, nextChunkLen int, err error) (bool, error) {
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read multipart upload data failed, %w", err)
	}

	if nextChunkLen == 0 {
		// No need to upload empty part, if file was empty to start
		// with empty single part would of been created and never
		// started multipart upload.
		return false, nil
	}

	part++
	// This upload exceeded maximum number of supported parts, error now.
	if part > u.cfg.MaxUploadParts || part > MaxUploadParts {
		var msg string
		if part > u.cfg.MaxUploadParts {
			msg = fmt.Sprintf("exceeded total allowed configured MaxUploadParts (%d). Adjust PartSize to fit in this limit",
				u.cfg.MaxUploadParts)
		} else {
			msg = fmt.Sprintf("exceeded total allowed S3 limit MaxUploadParts (%d). Adjust PartSize to fit in this limit",
				MaxUploadParts)
		}
		return false, fmt.Errorf(msg)
	}

	return true, err
}

// readChunk runs in worker goroutines to pull chunks off of the ch channel
// and send() them as UploadPart requests.
func (u *multiuploader) readChunk(ch chan chunk) {
	defer u.wg.Done()
	for {
		data, ok := <-ch

		if !ok {
			break
		}

		if eTag, present := u.eTagByPartNumber[data.num]; present {
			if err := u.check(data, &eTag); err != nil {
				u.seterr(err)
			}
		} else if u.geterr() == nil {
			if err := u.send(data); err != nil {
				u.seterr(err)
			}
		}

		data.cleanup()
	}
}

// completePart keeps track of completed part information
func (u *multiuploader) completePart(c chunk, eTag *string) {
	n := c.num
	completed := types.CompletedPart{ETag: eTag, PartNumber: n}

	u.m.Lock()
	u.parts = append(u.parts, completed)
	u.m.Unlock()
}

// check checks if a chunk's checksum matches its parts ETAG
// and keeps track of the completed part information
func (u *multiuploader) check(c chunk, eTag *string) error {
	summer := md5.New()
	io.Copy(summer, c.buf)
	sum := hex.EncodeToString(summer.Sum([]byte{}))
	if sum != *eTag {
		return fmt.Errorf("checksum did not match for chunk %d, multipart upload out of sync with local file", c.num)
	}

	u.completePart(c, eTag)
	return nil
}

// send performs an UploadPart request and keeps track of the completed
// part information.
func (u *multiuploader) send(c chunk) error {
	params := &s3.UploadPartInput{
		Bucket:               u.in.Bucket,
		Key:                  u.in.Key,
		Body:                 c.buf,
		UploadId:             &u.uploadID,
		SSECustomerAlgorithm: u.in.SSECustomerAlgorithm,
		SSECustomerKey:       u.in.SSECustomerKey,
		PartNumber:           c.num,
	}

	resp, err := u.cfg.S3.UploadPart(u.ctx, params, u.cfg.ClientOptions...)
	if err != nil {
		return err
	}

	u.completePart(c, resp.ETag)

	return nil
}

// geterr is a thread-safe getter for the error object
func (u *multiuploader) geterr() error {
	u.m.Lock()
	defer u.m.Unlock()

	return u.err
}

// seterr is a thread-safe setter for the error object
func (u *multiuploader) seterr(e error) {
	u.m.Lock()
	defer u.m.Unlock()

	u.err = e
}

// fail will abort the multipart unless LeavePartsOnError is set to true.
func (u *multiuploader) fail() {
	if u.cfg.LeavePartsOnError {
		return
	}

	params := &s3.AbortMultipartUploadInput{
		Bucket:   u.in.Bucket,
		Key:      u.in.Key,
		UploadId: &u.uploadID,
	}
	_, err := u.cfg.S3.AbortMultipartUpload(u.ctx, params, u.cfg.ClientOptions...)
	if err != nil {
		// TODO: Add logging
		//logMessage(u.cfg.S3, aws.LogDebug, fmt.Sprintf("failed to abort multipart upload, %v", err))
		_ = err
	}
}

// complete successfully completes a multipart upload and returns the response.
func (u *multiuploader) complete() *s3.CompleteMultipartUploadOutput {
	if u.geterr() != nil {
		u.fail()
		return nil
	}

	// Parts must be sorted in PartNumber order.
	sort.Sort(u.parts)

	params := &s3.CompleteMultipartUploadInput{
		Bucket:          u.in.Bucket,
		Key:             u.in.Key,
		UploadId:        &u.uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: u.parts},
	}
	resp, err := u.cfg.S3.CompleteMultipartUpload(u.ctx, params, u.cfg.ClientOptions...)
	if err != nil {
		u.seterr(err)
		u.fail()
	}

	return resp
}

type readerAtSeeker interface {
	io.ReaderAt
	io.ReadSeeker
}
