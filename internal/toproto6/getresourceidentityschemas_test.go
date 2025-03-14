// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGetResourceIdentitySchemasResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetResourceIdentitySchemasResponse
		expected *tfprotov6.GetResourceIdentitySchemasResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"resource-identity-identity-multiple-resources": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource_1": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.BoolAttribute{
								RequiredForImport: true,
							},
						},
					},
					"test_resource_2": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.BoolAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource_1": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								RequiredForImport: true,
								Name:              "test_attribute",
								Type:              tftypes.Bool,
							},
						},
					},
					"test_resource_2": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								RequiredForImport: true,
								Name:              "test_attribute",
								Type:              tftypes.Bool,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-optionalforimport": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.BoolAttribute{
								OptionalForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								OptionalForImport: true,
								Type:              tftypes.Bool,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-requiredforimport": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.BoolAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								Type:              tftypes.Bool,
								RequiredForImport: true,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-bool": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.BoolAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Bool,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-float32": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.Float32Attribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Number,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-float64": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.Float64Attribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Number,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-int32": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.Int32Attribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Number,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-int64": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.Int64Attribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Number,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-list-string": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.ListAttribute{
								RequiredForImport: true,
								ElementType:       types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type: tftypes.List{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-number": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.NumberAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.Number,
							},
						},
					},
				},
			},
		},
		"resource-identity-attribute-type-string": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Attributes: map[string]identityschema.Attribute{
							"test_attribute": identityschema.StringAttribute{
								RequiredForImport: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{
							{
								Name:              "test_attribute",
								RequiredForImport: true,
								Type:              tftypes.String,
							},
						},
					},
				},
			},
		},
		"resource-identity-version": {
			input: &fwserver.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]fwschema.Schema{
					"test_resource": identityschema.Schema{
						Version: 123,
					},
				},
			},
			expected: &tfprotov6.GetResourceIdentitySchemasResponse{
				IdentitySchemas: map[string]*tfprotov6.ResourceIdentitySchema{
					"test_resource": {
						IdentityAttributes: []*tfprotov6.ResourceIdentitySchemaAttribute{},
						Version:            123,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetResourceIdentitySchemasResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
