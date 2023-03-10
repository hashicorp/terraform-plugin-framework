package schema_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
		"AttributeName-block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"testblock": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			step: tftypes.AttributeName("testblock"),
			expected: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
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
		name, testCase := name, testCase

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
		"WithAttributeName-block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"other": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"otherattr": schema.StringAttribute{},
						},
					},
					"test": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test\n"+
						"Original Error: "+fwschema.ErrPathIsBlock.Error(),
				),
			},
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
		name, tc := name, tc

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
		"WithAttributeName-block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"other": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"otherattr": schema.StringAttribute{},
						},
					},
					"test": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test"),
			expected:    nil,
			expectedErr: fwschema.ErrPathIsBlock.Error(),
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
		name, tc := name, tc

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

			if err == nil && tc.expectedErr != "" {
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
		name, testCase := name, testCase

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
		"blocks": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"testblock1": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
					"testblock2": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			expected: map[string]fwschema.Block{
				"testblock1": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
				"testblock2": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

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
		name, testCase := name, testCase

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
		name, testCase := name, testCase

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
		name, testCase := name, testCase

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
		"version": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
				Version: 1,
			},
			expected: 1,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

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
				Blocks: map[string]schema.Block{
					"testblock": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
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
		name, testCase := name, testCase

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
		"AttributeName-Block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"list_block_nested": schema.StringAttribute{},
							},
						},
					},
					"set_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"set_block_nested": schema.StringAttribute{},
							},
						},
					},
					"single_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"single_block_nested": schema.StringAttribute{},
						},
					},
				},
			},
			path: path.Root("list_block"),
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"list_block_nested": types.StringType,
					},
				},
			},
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
		name, testCase := name, testCase

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
		"AttributeName-Block": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"list_block_nested": schema.StringAttribute{},
							},
						},
					},
					"set_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"set_block_nested": schema.StringAttribute{},
							},
						},
					},
					"single_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"single_block_nested": schema.StringAttribute{},
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("list_block"),
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"list_block_nested": types.StringType,
					},
				},
			},
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
		name, testCase := name, testCase

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

func TestSchemaValidateFieldName(t *testing.T) {
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
				diag.NewAttributeErrorDiagnostic(
					path.Root("depends_on"),
					"Schema Using Reserved Field Name",
					`"depends_on" is a reserved field name`,
				),
			},
		},
		"block-using-reserved-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"connection": schema.ListNestedBlock{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("connection"),
					"Schema Using Reserved Field Name",
					`"connection" is a reserved field name`,
				),
			},
		},
		"single-nested-attribute-using-nested-reserved-field-name": {
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
		"single-nested-block-using-nested-reserved-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"connection": schema.BoolAttribute{},
						},
					},
				},
			},
		},
		"list-nested-attribute-using-nested-reserved-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"depends_on": schema.Int64Attribute{},
							},
						},
					},
				},
			},
		},
		"list-nested-block-using-nested-reserved-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"connection": schema.BoolAttribute{},
							},
						},
					},
				},
			},
		},
		"attribute-and-blocks-using-reserved-field-names": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"depends_on": schema.StringAttribute{},
				},
				Blocks: map[string]schema.Block{
					"connection": schema.ListNestedBlock{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("depends_on"),
					"Schema Using Reserved Field Name",
					`"depends_on" is a reserved field name`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("connection"),
					"Schema Using Reserved Field Name",
					`"connection" is a reserved field name`,
				),
			},
		},
		"attribute-using-invalid-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"^": schema.StringAttribute{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"block-using-invalid-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"^": schema.ListNestedBlock{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"single-nested-attribute-using-nested-invalid-field-name": {
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
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_attribute").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"single-nested-block-using-nested-invalid-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"^": schema.BoolAttribute{},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_block").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"single-nested-attribute-using-invalid-field-names": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"$": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"^": schema.BoolAttribute{},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"single-nested-block-using-invalid-field-names": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"$": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"^": schema.BoolAttribute{},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"single-nested-block-with-nested-block-using-invalid-field-names": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"$": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"^": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"!": schema.BoolAttribute{},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^").AtName("!"),
					"Invalid Schema Field Name",
					`Field name "!" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"list-nested-attribute-using-nested-invalid-field-name": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"^": schema.Int64Attribute{},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_attribute").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"list-nested-block-using-nested-invalid-field-name": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"^": schema.Int64Attribute{},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_block").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"list-nested-attribute-using-invalid-field-names": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"$": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"^": schema.Int64Attribute{},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"list-nested-block-using-invalid-field-names": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"$": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"^": schema.Int64Attribute{},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
		"list-nested-block-with-nested-block-using-invalid-field-names": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"$": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"^": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"!": schema.BoolAttribute{},
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("$"),
					"Invalid Schema Field Name",
					`Field name "$" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^"),
					"Invalid Schema Field Name",
					`Field name "^" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("$").AtName("^").AtName("!"),
					"Invalid Schema Field Name",
					`Field name "!" is invalid, the only allowed characters are a-z, 0-9 and _. This is always a problem with the provider and should be reported to the provider developer.`,
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.schema.Validate()

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaValidateDefault(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        schema.Schema
		expectedDiags diag.Diagnostics
	}{
		"empty-schema": {
			schema: schema.Schema{},
		},
		"non-computed-bool-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Default: booldefault.StaticValue(true),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("bool_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "bool_attribute" must be computed when using default`,
				),
			},
		},
		"computed-bool-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"bool_attribute": schema.BoolAttribute{
						Computed: true,
						Default:  booldefault.StaticValue(true),
					},
				},
			},
		},
		"non-computed-float64-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Default: float64default.StaticValue(1.2345),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("float64_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "float64_attribute" must be computed when using default`,
				),
			},
		},
		"computed-float64-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"float64_attribute": schema.Float64Attribute{
						Computed: true,
						Default:  float64default.StaticValue(1.2345),
					},
				},
			},
		},
		"non-computed-int64-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Default: int64default.StaticValue(12345),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("int64_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "int64_attribute" must be computed when using default`,
				),
			},
		},
		"computed-int64-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"int64_attribute": schema.Int64Attribute{
						Computed: true,
						Default:  int64default.StaticValue(12345),
					},
				},
			},
		},
		"non-computed-list-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_attribute": schema.ListAttribute{
						Computed: true,
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-map-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_attribute": schema.MapAttribute{
						Default: mapdefault.StaticValue(
							types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "map_attribute" must be computed when using default`,
				),
			},
		},
		"computed-map-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map": schema.MapAttribute{
						Computed: true,
						Default: mapdefault.StaticValue(
							types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-number-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Default: numberdefault.StaticValue(types.NumberValue(big.NewFloat(1.2345))),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("number_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "number_attribute" must be computed when using default`,
				),
			},
		},
		"computed-number-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"number_attribute": schema.NumberAttribute{
						Computed: true,
						Default:  numberdefault.StaticValue(types.NumberValue(big.NewFloat(1.2345))),
					},
				},
			},
		},
		"non-computed-object-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"test-key": types.StringType,
								},
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "object_attribute" must be computed when using default`,
				),
			},
		},
		"computed-object-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"object_attribute": schema.ObjectAttribute{
						Computed: true,
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"test-key": types.StringType,
								},
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-set-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_attribute": schema.SetAttribute{
						Computed: true,
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Default: stringdefault.StaticValue("str"),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Computed: true,
						Default:  stringdefault.StaticValue("str"),
					},
				},
			},
		},
		"non-computed-list-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						Computed: true,
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-list-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"non-computed-list-nested-attribute-using-default-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
						Default: listdefault.StaticValue(
							types.ListValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
		},
		"non-computed-map-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Default: mapdefault.StaticValue(
							types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "map_nested_attribute" must be computed when using default`,
				),
			},
		},
		"computed-map-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Computed: true,
						Default: mapdefault.StaticValue(
							types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-map-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "map_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"non-computed-map-nested-attribute-using-default-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
						Default: mapdefault.StaticValue(
							types.MapValueMust(
								types.StringType,
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "map_nested_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("map_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "map_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-map-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
		},
		"non-computed-set-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						Computed: true,
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-set-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"non-computed-set-nested-attribute-using-default-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
		},
		"non-computed-single-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"test-key": types.StringType,
								},
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_attribute" must be computed when using default`,
				),
			},
		},
		"computed-single-nested-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Computed: true,
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"test-key": types.StringType,
								},
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
		},
		"non-computed-single-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Default: stringdefault.StaticValue("str"),
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"non-computed-single-nested-attribute-using-default-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Default: stringdefault.StaticValue("str"),
							},
						},
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"test-key": types.StringType,
								},
								map[string]attr.Value{
									"test-key": types.StringValue("str"),
								},
							),
						),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_attribute").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_attribute.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-single-nested-attribute-string-attribute-using-default": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
								Default:  stringdefault.StaticValue("str"),
							},
						},
					},
				},
			},
		},
		"non-computed-list-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
		},
		"non-computed-list-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"list_nested_nested_block": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Default: stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_block").AtName("list_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_block.list_nested_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"list_nested_nested_block": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Computed: true,
												Default:  stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"non-computed-list-nested-block-string-attribute-using-default-list-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
							Blocks: map[string]schema.Block{
								"list_nested_nested_block": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Default: stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_block").AtName("list_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_block.list_nested_nested_block.string_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("list_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "list_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-list-nested-block-string-attribute-using-default-list-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
							Blocks: map[string]schema.Block{
								"list_nested_nested_block": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Computed: true,
												Default:  stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"non-computed-set-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
						},
					},
				},
			},
		},
		"non-computed-set-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"set_nested_nested_block": schema.SetNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Default: stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_block").AtName("set_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_block.set_nested_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Blocks: map[string]schema.Block{
								"set_nested_nested_block": schema.SetNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Computed: true,
												Default:  stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"non-computed-set-nested-block-string-attribute-using-default-set-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Default: stringdefault.StaticValue("str"),
								},
							},
							Blocks: map[string]schema.Block{
								"set_nested_nested_block": schema.SetNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Default: stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_block").AtName("set_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_block.set_nested_nested_block.string_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("set_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "set_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-set-nested-block-string-attribute-using-default-set-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
									Default:  stringdefault.StaticValue("str"),
								},
							},
							Blocks: map[string]schema.Block{
								"set_nested_nested_block": schema.SetNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"string_attribute": schema.StringAttribute{
												Computed: true,
												Default:  stringdefault.StaticValue("str"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"non-computed-single-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Default: stringdefault.StaticValue("str"),
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-single-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
								Default:  stringdefault.StaticValue("str"),
							},
						},
					},
				},
			},
		},
		"non-computed-single-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"single_nested_nested_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Default: stringdefault.StaticValue("str"),
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_block").AtName("single_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_block.single_nested_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-single-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Blocks: map[string]schema.Block{
							"single_nested_nested_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
										Default:  stringdefault.StaticValue("str"),
									},
								},
							},
						},
					},
				},
			},
		},
		"non-computed-single-nested-block-string-attribute-using-default-single-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Default: stringdefault.StaticValue("str"),
							},
						},
						Blocks: map[string]schema.Block{
							"single_nested_nested_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Default: stringdefault.StaticValue("str"),
									},
								},
							},
						},
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_block").AtName("single_nested_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_block.single_nested_nested_block.string_attribute" must be computed when using default`,
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("single_nested_block").AtName("string_attribute"),
					"Schema Using Attribute Default For Non-Computed Attribute",
					`attribute "single_nested_block.string_attribute" must be computed when using default`,
				),
			},
		},
		"computed-single-nested-block-string-attribute-using-default-single-nested-nested-nested-block-string-attribute-using-default": {
			schema: schema.Schema{
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
								Default:  stringdefault.StaticValue("str"),
							},
						},
						Blocks: map[string]schema.Block{
							"single_nested_nested_block": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
										Default:  stringdefault.StaticValue("str"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.schema.Validate()

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+wanted, -got): %s", diff)
			}
		})
	}
}
