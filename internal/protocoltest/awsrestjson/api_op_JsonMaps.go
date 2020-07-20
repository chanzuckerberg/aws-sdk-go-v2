// Code generated by smithy-go-codegen DO NOT EDIT.

package awsrestjson

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/internal/protocoltest/awsrestjson/types"
	smithy "github.com/awslabs/smithy-go"
	"github.com/awslabs/smithy-go/middleware"
	smithyhttp "github.com/awslabs/smithy-go/transport/http"
)

// The example tests basic map serialization.
func (c *Client) JsonMaps(ctx context.Context, params *JsonMapsInput, optFns ...func(*Options)) (*JsonMapsOutput, error) {
	stack := middleware.NewStack("JsonMaps", smithyhttp.NewStackRequest)
	options := c.options.Copy()
	for _, fn := range optFns {
		fn(&options)
	}
	awsmiddleware.AddRequestInvocationIDMiddleware(stack)
	smithyhttp.AddContentLengthMiddleware(stack)
	AddResolveEndpointMiddleware(stack, options)
	retry.AddRetryMiddlewares(stack, options)
	awsmiddleware.AddAttemptClockSkewMiddleware(stack)
	stack.Initialize.Add(newServiceMetadataMiddleware_opJsonMaps(options.Region), middleware.Before)
	addawsRestjson1_serdeOpJsonMapsMiddlewares(stack)

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
			OperationName: "JsonMaps",
			Err:           err,
		}
	}
	out := result.(*JsonMapsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type JsonMapsInput struct {
	MyMap map[string]*types.GreetingStruct
}

type JsonMapsOutput struct {
	MyMap map[string]*types.GreetingStruct

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addawsRestjson1_serdeOpJsonMapsMiddlewares(stack *middleware.Stack) {
	stack.Serialize.Add(&awsRestjson1_serializeOpJsonMaps{}, middleware.After)
	stack.Deserialize.Add(&awsRestjson1_deserializeOpJsonMaps{}, middleware.After)
}

func newServiceMetadataMiddleware_opJsonMaps(region string) awsmiddleware.RegisterServiceMetadata {
	return awsmiddleware.RegisterServiceMetadata{
		Region:         region,
		ServiceName:    "Rest Json Protocol",
		ServiceID:      "restjsonprotocol",
		EndpointPrefix: "restjsonprotocol",
		OperationName:  "JsonMaps",
	}
}
