// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestGetMetadataResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetMetadataResponse
		expected *tfprotov5.GetMetadataResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"actions": {
			input: &fwserver.GetMetadataResponse{
				Actions: []fwserver.ActionMetadata{
					{
						TypeName: "test_action_1",
					},
					{
						TypeName: "test_action_2",
					},
				},
			},
			expected: &tfprotov5.GetMetadataResponse{
				Actions: []tfprotov5.ActionMetadata{
					{
						TypeName: "test_action_1",
					},
					{
						TypeName: "test_action_2",
					},
				},
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
			},
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
			expected: &tfprotov5.GetMetadataResponse{
				Actions: []tfprotov5.ActionMetadata{},
				DataSources: []tfprotov5.DataSourceMetadata{
					{
						TypeName: "test_data_source_1",
					},
					{
						TypeName: "test_data_source_2",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
			},
		},
		"ephemeralresources": {
			input: &fwserver.GetMetadataResponse{
				EphemeralResources: []fwserver.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource_1",
					},
					{
						TypeName: "test_ephemeral_resource_2",
					},
				},
			},
			expected: &tfprotov5.GetMetadataResponse{
				Actions:     []tfprotov5.ActionMetadata{},
				DataSources: []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource_1",
					},
					{
						TypeName: "test_ephemeral_resource_2",
					},
				},
				Functions:     []tfprotov5.FunctionMetadata{},
				ListResources: []tfprotov5.ListResourceMetadata{},
				Resources:     []tfprotov5.ResourceMetadata{},
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
			expected: &tfprotov5.GetMetadataResponse{
				Actions:     []tfprotov5.ActionMetadata{},
				DataSources: []tfprotov5.DataSourceMetadata{},
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "Duplicate Data Source Type Defined",
						Detail: "The test_data_source data source type name was returned for multiple data sources. " +
							"Data source type names must be unique. " +
							"This is always an issue with the provider and should be reported to the provider developers.",
					},
				},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
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
			expected: &tfprotov5.GetMetadataResponse{
				Actions:            []tfprotov5.ActionMetadata{},
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions: []tfprotov5.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				ListResources: []tfprotov5.ListResourceMetadata{},
				Resources:     []tfprotov5.ResourceMetadata{},
			},
		},
		"listresources": {
			input: &fwserver.GetMetadataResponse{
				ListResources: []fwserver.ListResourceMetadata{
					{
						TypeName: "test_list_resource_1",
					},
					{
						TypeName: "test_list_resource_2",
					},
				},
			},
			expected: &tfprotov5.GetMetadataResponse{
				Actions:            []tfprotov5.ActionMetadata{},
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources: []tfprotov5.ListResourceMetadata{
					{
						TypeName: "test_list_resource_1",
					},
					{
						TypeName: "test_list_resource_2",
					},
				},
				Resources: []tfprotov5.ResourceMetadata{},
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
			expected: &tfprotov5.GetMetadataResponse{
				Actions:            []tfprotov5.ActionMetadata{},
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				Resources: []tfprotov5.ResourceMetadata{
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
			expected: &tfprotov5.GetMetadataResponse{
				Actions:            []tfprotov5.ActionMetadata{},
				DataSources:        []tfprotov5.DataSourceMetadata{},
				EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
				Functions:          []tfprotov5.FunctionMetadata{},
				Resources:          []tfprotov5.ResourceMetadata{},
				ListResources:      []tfprotov5.ListResourceMetadata{},
				ServerCapabilities: &tfprotov5.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.GetMetadataResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
