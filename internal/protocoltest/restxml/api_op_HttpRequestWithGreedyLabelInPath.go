// Code generated by smithy-go-codegen DO NOT EDIT.

package restxml

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	smithy "github.com/awslabs/smithy-go"
	"github.com/awslabs/smithy-go/middleware"
	smithyhttp "github.com/awslabs/smithy-go/transport/http"
)

func (c *Client) HttpRequestWithGreedyLabelInPath(ctx context.Context, params *HttpRequestWithGreedyLabelInPathInput, optFns ...func(*Options)) (*HttpRequestWithGreedyLabelInPathOutput, error) {
	stack := middleware.NewStack("HttpRequestWithGreedyLabelInPath", smithyhttp.NewStackRequest)
	options := c.options.Copy()
	for _, fn := range optFns {
		fn(&options)
	}
	awsmiddleware.AddRequestInvocationIDMiddleware(stack)
	smithyhttp.AddContentLengthMiddleware(stack)
	AddResolveEndpointMiddleware(stack, options)
	retry.AddRetryMiddlewares(stack, options)
	awsmiddleware.AddAttemptClockSkewMiddleware(stack)
	addOpHttpRequestWithGreedyLabelInPathValidationMiddleware(stack)
	stack.Initialize.Add(newServiceMetadataMiddleware_opHttpRequestWithGreedyLabelInPath(options.Region), middleware.Before)
	addawsRestxml_serdeOpHttpRequestWithGreedyLabelInPathMiddlewares(stack)

	for _, fn := range options.APIOptions {
		if err := fn(stack); err != nil {
			return nil, err
		}
	}
	handler := middleware.DecorateHandler(smithyhttp.NewClientHandler(options.HTTPClient), stack)
	result, metadata, err := handler.Handle(ctx, params)
	if err != nil {
		return nil, &smithy.OperationError{
			ServiceID:     c.ServiceID(),
			OperationName: "HttpRequestWithGreedyLabelInPath",
			Err:           err,
		}
	}
	out := result.(*HttpRequestWithGreedyLabelInPathOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type HttpRequestWithGreedyLabelInPathInput struct {
	Foo *string
	Baz *string
}

type HttpRequestWithGreedyLabelInPathOutput struct {
	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addawsRestxml_serdeOpHttpRequestWithGreedyLabelInPathMiddlewares(stack *middleware.Stack) {
	stack.Serialize.Add(&awsRestxml_serializeOpHttpRequestWithGreedyLabelInPath{}, middleware.After)
	stack.Deserialize.Add(&awsRestxml_deserializeOpHttpRequestWithGreedyLabelInPath{}, middleware.After)
}

func newServiceMetadataMiddleware_opHttpRequestWithGreedyLabelInPath(region string) awsmiddleware.RegisterServiceMetadata {
	return awsmiddleware.RegisterServiceMetadata{
		Region:         region,
		ServiceName:    "Rest Xml Protocol",
		ServiceID:      "restxmlprotocol",
		EndpointPrefix: "restxmlprotocol",
		OperationName:  "HttpRequestWithGreedyLabelInPath",
	}
}
