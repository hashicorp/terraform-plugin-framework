// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/statestore/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      schema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-missing": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("could not find attribute or block \"other\" in schema"),
		},
		"ElementKeyInt": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to schema"),
		},
		"ElementKeyString": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to schema"),
		},
		"ElementKeyValue": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to schema"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.schema.ApplyTerraform5AttributePathStep(testCase.step)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaAttributeAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		path          path.Path
		expected      fwschema.Attribute
		expectedDiags diag.Diagnostics
	}{
		"empty-root": {
			schema:   schema.Schema{},
			path:     path.Empty(),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: \n"+
						"Original Error: got unexpected type schema.Schema",
				),
			},
		},
		"root": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:     path.Empty(),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: \n"+
						"Original Error: got unexpected type schema.Schema",
				),
			},
		},
		"WithAttributeName-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"other": schema.BoolAttribute{},
					"test":  schema.StringAttribute{},
				},
			},
			path:     path.Root("test"),
			expected: schema.StringAttribute{},
		},
		"WithElementKeyInt": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:     path.Empty().AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
				),
			},
		},
		"WithElementKeyString": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:     path.Empty().AtMapKey("test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [\"test\"]\n"+
						"Original Error: ElementKeyString(\"test\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
				),
			},
		},
		"WithElementKeyValue": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:     path.Empty().AtSetValue(types.StringValue("test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtSetValue(types.StringValue("test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [Value(\"test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"test\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
				),
			},
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := tc.schema.AttributeAtPath(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaAttributeAtTerraformPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema      schema.Schema
		path        *tftypes.AttributePath
		expected    fwschema.Attribute
		expectedErr string
	}{
		"empty-root": {
			schema:      schema.Schema{},
			path:        tftypes.NewAttributePath(),
			expected:    nil,
			expectedErr: "got unexpected type schema.Schema",
		},
		"empty-nil": {
			schema:      schema.Schema{},
			path:        nil,
			expected:    nil,
			expectedErr: "got unexpected type schema.Schema",
		},
		"root": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:        tftypes.NewAttributePath(),
			expected:    nil,
			expectedErr: "got unexpected type schema.Schema",
		},
		"nil": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:        nil,
			expected:    nil,
			expectedErr: "got unexpected type schema.Schema",
		},
		"WithAttributeName-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"other": schema.BoolAttribute{},
					"test":  schema.StringAttribute{},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: schema.StringAttribute{},
		},
		"WithElementKeyInt": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyInt(0),
			expected:    nil,
			expectedErr: "ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
		},
		"WithElementKeyString": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyString("test"),
			expected:    nil,
			expectedErr: "ElementKeyString(\"test\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
		},
		"WithElementKeyValue": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.StringAttribute{},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:    nil,
			expectedErr: "ElementKeyValue(tftypes.String<\"test\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
		},
	}

	for name, tc := range testCases {

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.schema.AttributeAtTerraformPath(context.Background(), tc.path)

			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}

			if tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaGetAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected map[string]fwschema.Attribute
	}{
		"no-attributes": {
			schema:   schema.Schema{},
			expected: map[string]fwschema.Attribute{},
		},
		"attributes": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr1": schema.StringAttribute{},
					"testattr2": schema.StringAttribute{},
				},
			},
			expected: map[string]fwschema.Attribute{
				"testattr1": schema.StringAttribute{},
				"testattr2": schema.StringAttribute{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetAttributes()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaGetBlocks(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected map[string]fwschema.Block
	}{
		"no-blocks": {
			schema:   schema.Schema{},
			expected: map[string]fwschema.Block{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetBlocks()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected string
	}{
		"no-deprecation-message": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"deprecation-message": {
			schema: schema.Schema{
				DeprecationMessage: "test deprecation message",
			},
			expected: "test deprecation message",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetDeprecationMessage()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected string
	}{
		"no-description": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"description": {
			schema: schema.Schema{
				Description: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected string
	}{
		"no-markdown-description": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"markdown-description": {
			schema: schema.Schema{
				MarkdownDescription: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetMarkdownDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaGetVersion(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected int64
	}{
		"no-version": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: 0,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.GetVersion()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   schema.Schema
		expected attr.Type
	}{
		"base": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"testattr": types.StringType,
					"testblock": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"testattr": types.StringType,
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.schema.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaTypeAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		path          path.Path
		expected      attr.Type
		expectedDiags diag.Diagnostics
	}{
		"empty-schema-empty-path": {
			schema:   schema.Schema{},
			path:     path.Empty(),
			expected: types.ObjectType{},
		},
		"empty-path": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool":   schema.BoolAttribute{},
					"string": schema.StringAttribute{},
				},
			},
			path: path.Empty(),
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"AttributeName-Attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool":   schema.BoolAttribute{},
					"string": schema.StringAttribute{},
				},
			},
			path:     path.Root("string"),
			expected: types.StringType,
		},
		"AttributeName-non-existent": {
			schema: schema.Schema{},
			path:   path.Root("non-existent"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("non-existent"),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: non-existent\n"+
						"Original Error: AttributeName(\"non-existent\") still remains in the path: could not find attribute or block \"non-existent\" in schema",
				),
			},
		},
		"ElementKeyInt": {
			schema: schema.Schema{},
			path:   path.Empty().AtListIndex(0),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
				),
			},
		},
		"ElementKeyString": {
			schema: schema.Schema{},
			path:   path.Empty().AtMapKey("invalid"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("invalid"),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [\"invalid\"]\n"+
						"Original Error: ElementKeyString(\"invalid\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
				),
			},
		},
		"ElementKeyValue": {
			schema: schema.Schema{},
			path:   path.Empty().AtSetValue(types.StringNull()),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtSetValue(types.StringNull()),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [Value(<null>)]\n"+
						"Original Error: ElementKeyValue(tftypes.String<null>) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testCase.schema.TypeAtPath(context.Background(), testCase.path)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaTypeAtTerraformPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		path          *tftypes.AttributePath
		expected      attr.Type
		expectedError error
	}{
		"empty-schema-nil-path": {
			schema:   schema.Schema{},
			path:     nil,
			expected: types.ObjectType{},
		},
		"empty-schema-empty-path": {
			schema:   schema.Schema{},
			path:     tftypes.NewAttributePath(),
			expected: types.ObjectType{},
		},
		"nil-path": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool":   schema.BoolAttribute{},
					"string": schema.StringAttribute{},
				},
			},
			path: nil,
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"empty-path": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool":   schema.BoolAttribute{},
					"string": schema.StringAttribute{},
				},
			},
			path: tftypes.NewAttributePath(),
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"AttributeName-Attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool":   schema.BoolAttribute{},
					"string": schema.StringAttribute{},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("string"),
			expected: types.StringType,
		},
		"AttributeName-non-existent": {
			schema:        schema.Schema{},
			path:          tftypes.NewAttributePath().WithAttributeName("non-existent"),
			expectedError: fmt.Errorf("AttributeName(\"non-existent\") still remains in the path: could not find attribute or block \"non-existent\" in schema"),
		},
		"ElementKeyInt": {
			schema:        schema.Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyInt(0),
			expectedError: fmt.Errorf("ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema"),
		},
		"ElementKeyString": {
			schema:        schema.Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyString("invalid"),
			expectedError: fmt.Errorf("ElementKeyString(\"invalid\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema"),
		},
		"ElementKeyValue": {
			schema:        schema.Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyValue(tftypes.NewValue(tftypes.String, nil)),
			expectedError: fmt.Errorf("ElementKeyValue(tftypes.String<null>) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.schema.TypeAtTerraformPath(context.Background(), testCase.path)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		expectedDiags diag.Diagnostics
	}{
		"empty-schema": {
			schema: schema.Schema{},
		},
		"attribute-using-reserved-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"depends_on": schema.StringAttribute{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"depends_on\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"nested-attribute-using-nested-reserved-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"depends_on": schema.BoolAttribute{},
						},
					},
				},
			},
		},
		"attribute-using-invalid-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"^": schema.StringAttribute{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"^\" at schema path \"^\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"nested-attribute-using-nested-invalid-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"^": schema.BoolAttribute{},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"^\" at schema path \"single_nested_attribute.^\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"attribute-with-validate-attribute-implementation-error": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"test": schema.ListAttribute{
						Required: true,
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute Implementation",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"test\" is missing the CustomType or ElementType field on a collection Attribute. "+
						"One of these fields is required to prevent other unexpected errors or panics.",
				),
			},
		},
		"nested-attribute-with-validate-attribute-implementation-error": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"test": schema.ListAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute Implementation",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"single_nested_attribute.test\" is missing the CustomType or ElementType field on a collection Attribute. "+
						"One of these fields is required to prevent other unexpected errors or panics.",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.schema.ValidateImplementation(context.Background())

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
