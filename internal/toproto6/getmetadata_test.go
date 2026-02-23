// Copyright IBM Corp. 2021, 2026
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
			expected: &tfprotov6.GetMetadataResponse{
				Actions: []tfprotov6.ActionMetadata{
					{
						TypeName: "test_action_1",
					},
					{
						TypeName: "test_action_2",
					},
				},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				StateStores:        []tfprotov6.StateStoreMetadata{},
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
			expected: &tfprotov6.GetMetadataResponse{
				Actions: []tfprotov6.ActionMetadata{},
				DataSources: []tfprotov6.DataSourceMetadata{
					{
						TypeName: "test_data_source_1",
					},
					{
						TypeName: "test_data_source_2",
					},
				},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				StateStores:        []tfprotov6.StateStoreMetadata{},
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
				Actions:     []tfprotov6.ActionMetadata{},
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
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				StateStores:        []tfprotov6.StateStoreMetadata{},
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
			expected: &tfprotov6.GetMetadataResponse{
				Actions:     []tfprotov6.ActionMetadata{},
				DataSources: []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{
					{
						TypeName: "test_ephemeral_resource_1",
					},
					{
						TypeName: "test_ephemeral_resource_2",
					},
				},
				Functions:     []tfprotov6.FunctionMetadata{},
				ListResources: []tfprotov6.ListResourceMetadata{},
				Resources:     []tfprotov6.ResourceMetadata{},
				StateStores:   []tfprotov6.StateStoreMetadata{},
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
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions: []tfprotov6.FunctionMetadata{
					{
						Name: "function1",
					},
					{
						Name: "function2",
					},
				},
				ListResources: []tfprotov6.ListResourceMetadata{},
				Resources:     []tfprotov6.ResourceMetadata{},
				StateStores:   []tfprotov6.StateStoreMetadata{},
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
			expected: &tfprotov6.GetMetadataResponse{
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources: []tfprotov6.ListResourceMetadata{
					{
						TypeName: "test_list_resource_1",
					},
					{
						TypeName: "test_list_resource_2",
					},
				},
				Resources:   []tfprotov6.ResourceMetadata{},
				StateStores: []tfprotov6.StateStoreMetadata{},
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
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources: []tfprotov6.ResourceMetadata{
					{
						TypeName: "test_resource_1",
					},
					{
						TypeName: "test_resource_2",
					},
				},
				StateStores: []tfprotov6.StateStoreMetadata{},
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
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
				ServerCapabilities: &tfprotov6.ServerCapabilities{
					GetProviderSchemaOptional: true,
					PlanDestroy:               true,
				},
				StateStores: []tfprotov6.StateStoreMetadata{},
			},
		},
		"statestores": {
			input: &fwserver.GetMetadataResponse{
				StateStores: []fwserver.StateStoreMetadata{
					{
						TypeName: "test_state_store_1",
					},
					{
						TypeName: "test_state_store_2",
					},
				},
			},
			expected: &tfprotov6.GetMetadataResponse{
				StateStores: []tfprotov6.StateStoreMetadata{
					{
						TypeName: "test_state_store_1",
					},
					{
						TypeName: "test_state_store_2",
					},
				},
				Actions:            []tfprotov6.ActionMetadata{},
				DataSources:        []tfprotov6.DataSourceMetadata{},
				EphemeralResources: []tfprotov6.EphemeralResourceMetadata{},
				Functions:          []tfprotov6.FunctionMetadata{},
				ListResources:      []tfprotov6.ListResourceMetadata{},
				Resources:          []tfprotov6.ResourceMetadata{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetMetadataResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
