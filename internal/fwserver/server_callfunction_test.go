// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/fwerror"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestServerCallFunction(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.CallFunctionRequest
		expectedResponse *fwserver.CallFunctionResponse
	}{
		"request-nil": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			expectedResponse: &fwserver.CallFunctionResponse{},
		},
		"request-arguments-get": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewBoolNull(),
					basetypes.NewInt64Unknown(),
					basetypes.NewStringValue("arg2"),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 basetypes.BoolValue
						var arg1 basetypes.Int64Value
						var arg2 basetypes.StringValue

						resp.Error.Append(req.Arguments.Get(ctx, &arg0, &arg1, &arg2)...)

						expectedArg0 := basetypes.NewBoolNull()
						expectedArg1 := basetypes.NewInt64Unknown()
						expectedArg2 := basetypes.NewStringValue("arg2")

						if !arg0.Equal(expectedArg0) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if !arg1.Equal(expectedArg1) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg1, expectedArg1)))
						}

						if !arg2.Equal(expectedArg2) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 2 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg2, expectedArg2)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"request-arguments-get-reflection": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewStringValue("arg0"),
					basetypes.NewStringNull(),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 string
						var arg1 *string

						resp.Error.Append(req.Arguments.Get(ctx, &arg0, &arg1)...)

						expectedArg0 := "arg0"

						if arg0 != expectedArg0 {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if arg1 != nil {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: nil", *arg1)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"request-arguments-get-variadic": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewStringValue("arg0"),
					basetypes.NewListValueMust(
						basetypes.StringType{},
						[]attr.Value{
							basetypes.NewStringValue("arg1-element0"),
							basetypes.NewStringValue("arg1-element1"),
						},
					),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 basetypes.StringValue
						var arg1 basetypes.ListValue

						resp.Error.Append(req.Arguments.Get(ctx, &arg0, &arg1)...)

						expectedArg0 := basetypes.NewStringValue("arg0")
						expectedArg1 := basetypes.NewListValueMust(
							basetypes.StringType{},
							[]attr.Value{
								basetypes.NewStringValue("arg1-element0"),
								basetypes.NewStringValue("arg1-element1"),
							},
						)

						if !arg0.Equal(expectedArg0) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if !arg1.Equal(expectedArg1) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg1, expectedArg1)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"request-arguments-getargument": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewBoolNull(),
					basetypes.NewInt64Unknown(),
					basetypes.NewStringValue("arg2"),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 basetypes.BoolValue
						var arg1 basetypes.Int64Value
						var arg2 basetypes.StringValue

						resp.Error.Append(req.Arguments.GetArgument(ctx, 0, &arg0)...)
						resp.Error.Append(req.Arguments.GetArgument(ctx, 1, &arg1)...)
						resp.Error.Append(req.Arguments.GetArgument(ctx, 2, &arg2)...)

						expectedArg0 := basetypes.NewBoolNull()
						expectedArg1 := basetypes.NewInt64Unknown()
						expectedArg2 := basetypes.NewStringValue("arg2")

						if !arg0.Equal(expectedArg0) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if !arg1.Equal(expectedArg1) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg1, expectedArg1)))
						}

						if !arg2.Equal(expectedArg2) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 2 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg2, expectedArg2)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"request-arguments-getargument-reflection": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewStringValue("arg0"),
					basetypes.NewStringNull(),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 string
						var arg1 *string

						resp.Error.Append(req.Arguments.GetArgument(ctx, 0, &arg0)...)
						resp.Error.Append(req.Arguments.GetArgument(ctx, 1, &arg1)...)

						expectedArg0 := "arg0"

						if arg0 != expectedArg0 {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if arg1 != nil {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: nil", *arg1)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"request-arguments-getargument-variadic": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData([]attr.Value{
					basetypes.NewStringValue("arg0"),
					basetypes.NewListValueMust(
						basetypes.StringType{},
						[]attr.Value{
							basetypes.NewStringValue("arg1-element0"),
							basetypes.NewStringValue("arg1-element1"),
						},
					),
				}),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						var arg0 basetypes.StringValue
						var arg1 basetypes.ListValue

						resp.Error.Append(req.Arguments.GetArgument(ctx, 0, &arg0)...)
						resp.Error.Append(req.Arguments.GetArgument(ctx, 1, &arg1)...)

						expectedArg0 := basetypes.NewStringValue("arg0")
						expectedArg1 := basetypes.NewListValueMust(
							basetypes.StringType{},
							[]attr.Value{
								basetypes.NewStringValue("arg1-element0"),
								basetypes.NewStringValue("arg1-element1"),
							},
						)

						if !arg0.Equal(expectedArg0) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 0 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg0, expectedArg0)))
						}

						if !arg1.Equal(expectedArg1) {
							resp.Error.Append(fwerror.NewErrorFunctionError("Unexpected Argument 1 Difference\n\n%s", fmt.Sprintf("got: %s, expected: %s", arg1, expectedArg1)))
						}

						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"response-diagnostics": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData(nil),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						resp.Error.Append(fwerror.NewWarningFunctionError("warning summary", "warning detail"))
						resp.Error.Append(fwerror.NewErrorFunctionError("error summary", "error detail"))
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error: fwerror.FunctionErrors{
					fwerror.NewWarningFunctionError("warning summary", "warning detail"),
					fwerror.NewErrorFunctionError("error summary", "error detail"),
				},
				Result: function.NewResultData(basetypes.NewStringUnknown()),
			},
		},
		"response-result": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData(nil),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
		"response-result-reflection": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{},
			},
			request: &fwserver.CallFunctionRequest{
				Arguments: function.NewArgumentsData(nil),
				Function: &testprovider.Function{
					RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
						resp.Error.Append(resp.Result.Set(ctx, "result")...)
					},
				},
				FunctionDefinition: function.Definition{
					Return: function.StringReturn{},
				},
			},
			expectedResponse: &fwserver.CallFunctionResponse{
				Error:  nil,
				Result: function.NewResultData(basetypes.NewStringValue("result")),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.CallFunctionResponse{}
			testCase.server.CallFunction(context.Background(), testCase.request, response)

			// Handling error comparison
			equateErrors := cmp.Comparer(func(x, y error) bool {
				if x == nil || y == nil {
					return x == nil && y == nil
				}

				return x.Error() == y.Error()
			})

			if diff := cmp.Diff(response, testCase.expectedResponse, equateErrors, cmp.AllowUnexported(function.ResultData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
