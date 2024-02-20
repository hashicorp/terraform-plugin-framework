// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestServerCallFunction(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov6.CallFunctionRequest
		expectedError    error
		expectedResponse *tfprotov6.CallFunctionResponse
	}{
		"request-arguments": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(ctx context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction"
										},
										DefinitionMethod: func(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Parameters: []function.Parameter{
													function.BoolParameter{},
													function.Int64Parameter{},
													function.StringParameter{},
												},
												Return: function.StringReturn{},
											}
										},
										RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
											var arg0 basetypes.BoolValue
											var arg1 basetypes.Int64Value
											var arg2 basetypes.StringValue

											resp.Error = req.Arguments.Get(ctx, &arg0, &arg1, &arg2)

											expectedArg0 := basetypes.NewBoolNull()
											expectedArg1 := basetypes.NewInt64Unknown()
											expectedArg2 := basetypes.NewStringValue("arg2")

											if !arg0.Equal(expectedArg0) {
												resp.Error.AddError(fmt.Sprintf("Unexpected Argument 0 Difference: got: %s, expected: %s", arg0, expectedArg0))
											}

											if !arg1.Equal(expectedArg1) {
												resp.Error.AddError(fmt.Sprintf("Unexpected Argument 1 Difference: got: %s, expected: %s", arg1, expectedArg1))
											}

											if !arg2.Equal(expectedArg2) {
												resp.Error.AddError(fmt.Sprintf("Unexpected Argument 2 Difference: got: %s, expected: %s", arg2, expectedArg2))
											}

											resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.CallFunctionRequest{
				Arguments: []*tfprotov6.DynamicValue{
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.Bool, nil)),
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)),
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "arg2")),
				},
				Name: "testfunction",
			},
			expectedResponse: &tfprotov6.CallFunctionResponse{
				Result: testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "result")),
			},
		},
		"request-arguments-variadic": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(ctx context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction"
										},
										DefinitionMethod: func(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Parameters: []function.Parameter{
													function.StringParameter{},
												},
												VariadicParameter: function.StringParameter{},
												Return:            function.StringReturn{},
											}
										},
										RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
											var arg0 basetypes.StringValue
											var arg1 basetypes.ListValue

											resp.Error = req.Arguments.Get(ctx, &arg0, &arg1)

											expectedArg0 := basetypes.NewStringValue("arg0")
											expectedArg1 := basetypes.NewListValueMust(
												basetypes.StringType{},
												[]attr.Value{
													basetypes.NewStringValue("varg-arg1"),
													basetypes.NewStringValue("varg-arg2"),
												},
											)

											if !arg0.Equal(expectedArg0) {
												resp.Error.AddError(fmt.Sprintf("Unexpected Argument 0 Difference: got: %s, expected: %s", arg0, expectedArg0))
											}

											if !arg1.Equal(expectedArg1) {
												resp.Error.AddError(fmt.Sprintf("Unexpected Argument 0 Difference: got: %s, expected: %s", arg1, expectedArg1))
											}

											resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.CallFunctionRequest{
				Arguments: []*tfprotov6.DynamicValue{
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "arg0")),
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "varg-arg1")),
					testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "varg-arg2")),
				},
				Name: "testfunction",
			},
			expectedResponse: &tfprotov6.CallFunctionResponse{
				Result: testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "result")),
			},
		},
		"response-function-errors": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(ctx context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction"
										},
										DefinitionMethod: func(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
											resp.Error.AddError("error summary: error detail")
											resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.CallFunctionRequest{
				Arguments: []*tfprotov6.DynamicValue{},
				Name:      "testfunction",
			},
			expectedResponse: &tfprotov6.CallFunctionResponse{
				Error: &tfprotov6.FunctionError{
					Text: "error summary: error detail\n",
				},
				Result: testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "result")),
			},
		},
		"response-result": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(ctx context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction"
										},
										DefinitionMethod: func(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										RunMethod: func(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
											resp.Error.Append(resp.Result.Set(ctx, basetypes.NewStringValue("result"))...)
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov6.CallFunctionRequest{
				Arguments: []*tfprotov6.DynamicValue{},
				Name:      "testfunction",
			},
			expectedResponse: &tfprotov6.CallFunctionResponse{
				Result: testNewSingleValueDynamicValue(t, tftypes.NewValue(tftypes.String, "result")),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.CallFunction(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
