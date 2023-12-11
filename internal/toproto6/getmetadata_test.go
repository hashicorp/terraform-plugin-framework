// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestGetMetadataResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetMetadataResponse
		expected *tfprotov6.GetMetadataResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"datasources": {
			input: &fwserver.GetMetadataResponse{
				DataSources: []fwserver.DataSourceMetadata{
					{
						TypeName: "test_data_source_1",
					},
					{
						TypeName: "test_data_source_2",
					},
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{
					{
						TypeName: "test_data_source_1",
					},
					{
						TypeName: "test_data_source_2",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
			},
		},
		"diagnostics": {
			input: &fwserver.GetMetadataResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Duplicate Data Source Type Defined",
						"The test_data_source data source type name was returned for multiple data sources. "+
							"Data source type names must be unique. "+
							"This is always an issue with the provider and should be reported to the provider developers.",
					),
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				Functions: []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{},
			},
		},
		"functions": {
			input: &fwserver.GetMetadataResponse{
				Functions: []fwserver.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Functions: []tfprotov6.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				Resources: []tfprotov6.ResourceMetadata{},
			},
		},
		"resources": {
			input: &fwserver.GetMetadataResponse{
				Resources: []fwserver.ResourceMetadata{
					{
						TypeName: "test_resource_1",
					},
					{
						TypeName: "test_resource_2",
					},
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Functions:   []tfprotov6.FunctionMetadata{},
				Resources: []tfprotov6.ResourceMetadata{
					{
						TypeName: "test_resource_1",
					},
					{
						TypeName: "test_resource_2",
					},
				},
			},
		},
		"servercapabilities": {
			input: &fwserver.GetMetadataResponse{
				ServerCapabilities: &fwserver.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				DataSources: []tfprotov6.DataSourceMetadata{},
				Functions:   []tfprotov6.FunctionMetadata{},
				Resources:   []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetMetadataResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
