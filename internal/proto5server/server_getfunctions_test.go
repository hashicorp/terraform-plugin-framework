// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerGetFunctions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *Server
		request          *tfprotov5.GetFunctionsRequest
		expectedError    error
		expectedResponse *tfprotov5.GetFunctionsResponse
	}{
		"functions": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function1"
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "function2"
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetFunctionsRequest{},
			expectedResponse: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"function1": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
					},
					"function2": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-duplicate-type-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
								func() function.Function {
									return &testprovider.Function{
										DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
											resp.Definition = function.Definition{
												Return: function.StringReturn{},
											}
										},
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "testfunction" // intentionally duplicate
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetFunctionsRequest{},
			expectedResponse: &tfprotov5.GetFunctionsResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Function Name Defined",
						Detail: "The testfunction function name was returned for multiple functions. " +
							"Function names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
			},
		},
		"functions-empty-name": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.ProviderWithFunctions{
						FunctionsMethod: func(_ context.Context) []func() function.Function {
							return []func() function.Function{
								func() function.Function {
									return &testprovider.Function{
										MetadataMethod: func(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
											resp.Name = "" // intentionally empty
										},
									}
								},
							}
						},
					},
				},
			},
			request: &tfprotov5.GetFunctionsRequest{},
			expectedResponse: &tfprotov5.GetFunctionsResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Function Name Missing",
						Detail: "The *testprovider.Function Function returned an empty string from the Metadata method. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.server.GetFunctions(context.Background(), new(tfprotov5.GetFunctionsRequest))

			if diff := cmp.Diff(testCase.expectedError, err); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedResponse, got); diff != "" {
				t.Errorf("unexpected response difference: %s", diff)
			}
		})
	}
}
