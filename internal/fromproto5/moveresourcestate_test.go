// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func TestMoveResourceStateRequest(t *testing.T) {
	t.Parallel()

	testFwSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov5.MoveResourceStateRequest
		resourceSchema      fwschema.Schema
		resource            resource.Resource
		expected            *fwserver.MoveResourceStateRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"SourcePrivate": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourcePrivate: privatestate.MustMarshalToJson(map[string][]byte{
					".frameworkKey":  []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourcePrivate: &privatestate.Data{
					Framework: map[string][]byte{
						".frameworkKey": []byte(`{"fKeyOne": {"k0": "zero", "k1": 1}}`),
					},
					Provider: privatestate.MustProviderData(context.Background(), privatestate.MustMarshalToJson(map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
					})),
				},
				TargetResourceSchema: testFwSchema,
			},
		},
		"SourcePrivate-malformed-json": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourcePrivate: []byte(`{`),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				TargetResourceSchema: testFwSchema,
			},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when decoding private state: unexpected end of JSON input.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"SourcePrivate-empty-json": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourcePrivate: []byte("{}"),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourcePrivate: &privatestate.Data{
					Framework: map[string][]byte{},
					Provider:  privatestate.EmptyProviderData(context.Background()),
				},
				TargetResourceSchema: testFwSchema,
			},
		},
		"SourceProviderAddress": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourceProviderAddress: "example.com/namespace/type",
				TargetResourceSchema:  testFwSchema,
			},
		},
		"SourceRawState": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourceState: testNewTfprotov5RawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourceRawState: testNewTfprotov6RawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
				TargetResourceSchema: testFwSchema,
			},
		},
		"SourceSchemaVersion": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourceSchemaVersion: 123,
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourceSchemaVersion:  123,
				TargetResourceSchema: testFwSchema,
			},
		},
		"SourceTypeName": {
			input: &tfprotov5.MoveResourceStateRequest{
				SourceTypeName: "examplecloud_thing",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				SourceTypeName:       "examplecloud_thing",
				TargetResourceSchema: testFwSchema,
			},
		},
		"TargetResourceSchema": {
			input:          &tfprotov5.MoveResourceStateRequest{},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				TargetResourceSchema: testFwSchema,
			},
		},
		"TargetResourceSchema-missing": {
			input:    &tfprotov5.MoveResourceStateRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Framework Implementation Error",
					"An unexpected issue was encountered when converting the MoveResourceState RPC request information from the protocol type to the framework type. "+
						"The resource schema was missing. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.",
				),
			},
		},
		"TargetTypeName": {
			input: &tfprotov5.MoveResourceStateRequest{
				TargetTypeName: "examplecloud_thing",
			},
			resourceSchema: testFwSchema,
			expected: &fwserver.MoveResourceStateRequest{
				TargetResourceSchema: testFwSchema,
				TargetTypeName:       "examplecloud_thing",
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.MoveResourceStateRequest(context.Background(), testCase.input, testCase.resource, testCase.resourceSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
