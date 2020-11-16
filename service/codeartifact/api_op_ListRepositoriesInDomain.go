// Code generated by smithy-go-codegen DO NOT EDIT.

package codeartifact

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/awslabs/smithy-go/middleware"
	smithyhttp "github.com/awslabs/smithy-go/transport/http"
)

// Returns a list of RepositorySummary
// (https://docs.aws.amazon.com/codeartifact/latest/APIReference/API_RepositorySummary.html)
// objects. Each RepositorySummary contains information about a repository in the
// specified domain and that matches the input parameters.
func (c *Client) ListRepositoriesInDomain(ctx context.Context, params *ListRepositoriesInDomainInput, optFns ...func(*Options)) (*ListRepositoriesInDomainOutput, error) {
	if params == nil {
		params = &ListRepositoriesInDomainInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListRepositoriesInDomain", params, optFns, addOperationListRepositoriesInDomainMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListRepositoriesInDomainOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListRepositoriesInDomainInput struct {

	// The name of the domain that contains the returned list of repositories.
	//
	// This member is required.
	Domain *string

	// Filter the list of repositories to only include those that are managed by the
	// AWS account ID.
	AdministratorAccount *string

	// The 12-digit account number of the AWS account that owns the domain. It does not
	// include dashes or spaces.
	DomainOwner *string

	// The maximum number of results to return per page.
	MaxResults *int32

	// The token for the next set of results. Use the value returned in the previous
	// response in the next request to retrieve the next set of results.
	NextToken *string

	// A prefix used to filter returned repositories. Only repositories with names that
	// start with repositoryPrefix are returned.
	RepositoryPrefix *string
}

type ListRepositoriesInDomainOutput struct {

	// If there are additional results, this is the token for the next set of results.
	NextToken *string

	// The returned list of repositories.
	Repositories []types.RepositorySummary

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addOperationListRepositoriesInDomainMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListRepositoriesInDomain{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListRepositoriesInDomain{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddAttemptClockSkewMiddleware(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpListRepositoriesInDomainValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListRepositoriesInDomain(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opListRepositoriesInDomain(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "codeartifact",
		OperationName: "ListRepositoriesInDomain",
	}
}