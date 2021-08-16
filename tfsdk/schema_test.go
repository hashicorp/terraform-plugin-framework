package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaAttributeType(t *testing.T) {
	testSchema := Schema{
		Attributes: map[string]Attribute{
			"foo": {
				Type:     types.StringType,
				Required: true,
			},
			"bar": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"disks": {
				Attributes: ListNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type:     types.BoolType,
						Optional: true,
					},
				}, ListNestedAttributesOptions{}),
				Optional: true,
				Computed: true,
			},
			"boot_disk": {
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type: types.BoolType,
					},
				}),
			},
		},
	}

	expectedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"foo": types.StringType,
			"bar": types.ListType{
				ElemType: types.StringType,
			},
			"disks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"delete_with_instance": types.BoolType,
					},
				},
			},
			"boot_disk": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
			},
		},
	}

	actualType := testSchema.AttributeType()

	if !expectedType.Equal(actualType) {
		t.Fatalf("types not equal (+wanted, -got): %s", cmp.Diff(expectedType, actualType))
	}
}

func TestSchemaTfprotov6Schema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Schema
		expected    *tfprotov6.Schema
		expectedErr string
	}

	tests := map[string]testCase{
		"empty-val": {
			input:       Schema{},
			expectedErr: "must have at least one attribute in the schema",
		},
		"basic-attrs": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
					"number": {
						Type:     types.NumberType,
						Optional: true,
					},
					"bool": {
						Type:     types.BoolType,
						Computed: true,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bool",
							Type:     tftypes.Bool,
							Computed: true,
						},
						{
							Name:     "number",
							Type:     tftypes.Number,
							Optional: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
				},
			},
		},
		"complex-attrs": {
			input: Schema{
				Version: 2,
				Attributes: map[string]Attribute{
					"list": {
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
					"object": {
						Type: types.ObjectType{AttrTypes: map[string]attr.Type{
							"string": types.StringType,
							"number": types.NumberType,
							"bool":   types.BoolType,
						}},
						Optional: true,
					},
					"map": {
						Type:     types.MapType{ElemType: types.NumberType},
						Computed: true,
					},
					// TODO: add tuple support when it lands
					// TODO: add set support when it lands
				},
			},
			expected: &tfprotov6.Schema{
				Version: 2,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.String},
							Required: true,
						},
						{
							Name:     "map",
							Type:     tftypes.Map{AttributeType: tftypes.Number},
							Computed: true,
						},
						{
							Name: "object",
							Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
								"string": tftypes.String,
								"number": tftypes.Number,
								"bool":   tftypes.Bool,
							}},
							Optional: true,
						},
					},
				},
			},
		},
		"nested-attrs": {
			input: Schema{
				Version: 3,
				Attributes: map[string]Attribute{
					"single": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}),
						Required: true,
					},
					"list": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, ListNestedAttributesOptions{}),
						Optional: true,
					},
					"set": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, SetNestedAttributesOptions{}),
						Computed: true,
					},
					"map": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, MapNestedAttributesOptions{}),
						Optional: true,
						Computed: true,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 3,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name: "list",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeList,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
						},
						{
							Name: "map",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeMap,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						{
							Name: "set",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSet,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Computed: true,
						},
						{
							Name: "single",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSingle,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
		},
		"markdown-description": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				MarkdownDescription: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindMarkdown,
				},
			},
		},
		"plaintext-description": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				Description: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindPlain,
				},
			},
		},
		"deprecated": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				DeprecationMessage: "deprecated, use other_resource instead",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Deprecated: true,
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.input.tfprotov6Schema(context.Background())
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
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestSchemaValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateSchemaRequest
		resp ValidateSchemaResponse
	}{
		"no-validation": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{},
		},
		"deprecation-message": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
							},
						},
						DeprecationMessage: "Use something else instead.",
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "Deprecated",
						Detail:   "Use something else instead.",
					},
				},
			},
		},
		"warnings": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testWarningDiagnostic,
					testWarningDiagnostic,
				},
			},
		},
		"errors": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testErrorDiagnostic,
					testErrorDiagnostic,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateSchemaResponse
			tc.req.Config.Schema.validate(context.Background(), tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
