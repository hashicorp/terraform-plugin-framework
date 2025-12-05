// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
)

func TestServerGetFunctions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.GetFunctionsRequest
		expectedResponse *fwserver.GetFunctionsResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{},
			},
		},
		"functiondefinitions": {
			server: &fwserver.Server{
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
			request: &fwserver.GetFunctionsRequest{},
			expectedResponse: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"function1": {
						Return: function.StringReturn{},
					},
					"function2": {
						Return: function.StringReturn{},
					},
				},
			},
		},
		"functiondefinitions-invalid-definition": {
			server: &fwserver.Server{
				Provider: &testprovider.ProviderWithFunctions{
					FunctionsMethod: func(_ context.Context) []func() function.Function {
						return []func() function.Function{
							func() function.Function {
								return &testprovider.Function{
									DefinitionMethod: func(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
										resp.Definition = function.Definition{
											Return: nil, // intentional
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
			request: &fwserver.GetFunctionsRequest{},
			expectedResponse: &fwserver.GetFunctionsResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"function1\" - Definition Return field is undefined",
					),
				},
				FunctionDefinitions: map[string]function.Definition{},
			},
		},
		"functiondefinitions-duplicate-type-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetFunctionsRequest{},
			expectedResponse: &fwserver.GetFunctionsResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Function Name Defined",
						"The testfunction function name was returned for multiple functions. "+
							"Function names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: map[string]function.Definition{},
			},
		},
		"functiondefinitions-empty-name": {
			server: &fwserver.Server{
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
			request: &fwserver.GetFunctionsRequest{},
			expectedResponse: &fwserver.GetFunctionsResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Function Name Missing",
						"The *testprovider.Function Function returned an empty string from the Metadata method. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
				FunctionDefinitions: map[string]function.Definition{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.GetFunctionsResponse{}
			testCase.server.GetFunctions(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
