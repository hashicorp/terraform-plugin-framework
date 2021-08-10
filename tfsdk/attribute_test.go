package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributeTfprotov6SchemaAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        Attribute
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaAttribute
		expectedErr string
	}

	tests := map[string]testCase{
		"deprecated": {
			name: "string",
			attr: Attribute{
				Type:               types.StringType,
				Optional:           true,
				DeprecationMessage: "deprecated, use new_string instead",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:       "string",
				Type:       tftypes.String,
				Optional:   true,
				Deprecated: true,
			},
		},
		"description-plain": {
			name: "string",
			attr: Attribute{
				Type:        types.StringType,
				Optional:    true,
				Description: "A string attribute",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute",
				DescriptionKind: tfprotov6.StringKindPlain,
			},
		},
		"description-markdown": {
			name: "string",
			attr: Attribute{
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "A string attribute",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
		},
		"description-both": {
			name: "string",
			attr: Attribute{
				Type:                types.StringType,
				Optional:            true,
				Description:         "A string attribute",
				MarkdownDescription: "A string attribute (markdown)",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute (markdown)",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
		},
		"attr-string": {
			name: "string",
			attr: Attribute{
				Type:     types.StringType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
		},
		"attr-bool": {
			name: "bool",
			attr: Attribute{
				Type:     types.BoolType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "bool",
				Type:     tftypes.Bool,
				Optional: true,
			},
		},
		"attr-number": {
			name: "number",
			attr: Attribute{
				Type:     types.NumberType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "number",
				Type:     tftypes.Number,
				Optional: true,
			},
		},
		"attr-list": {
			name: "list",
			attr: Attribute{
				Type:     types.ListType{ElemType: types.NumberType},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list",
				Type:     tftypes.List{ElementType: tftypes.Number},
				Optional: true,
			},
		},
		"attr-map": {
			name: "map",
			attr: Attribute{
				Type:     types.MapType{ElemType: types.StringType},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "map",
				Type:     tftypes.Map{AttributeType: tftypes.String},
				Optional: true,
			},
		},
		"attr-object": {
			name: "object",
			attr: Attribute{
				Type: types.ObjectType{AttrTypes: map[string]attr.Type{
					"foo": types.StringType,
					"bar": types.NumberType,
					"baz": types.BoolType,
				}},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name: "object",
				Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
					"baz": tftypes.Bool,
				}},
				Optional: true,
			},
		},
		// TODO: add set attribute when we support it
		// TODO: add tuple attribute when we support it
		"required": {
			name: "string",
			attr: Attribute{
				Type:     types.StringType,
				Required: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Required: true,
			},
		},
		"optional": {
			name: "string",
			attr: Attribute{
				Type:     types.StringType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
		},
		"computed": {
			name: "string",
			attr: Attribute{
				Type:     types.StringType,
				Computed: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Computed: true,
			},
		},
		"optional-computed": {
			name: "string",
			attr: Attribute{
				Type:     types.StringType,
				Computed: true,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Computed: true,
				Optional: true,
			},
		},
		"sensitive": {
			name: "string",
			attr: Attribute{
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:      "string",
				Type:      tftypes.String,
				Optional:  true,
				Sensitive: true,
			},
		},
		"nested-attr-single": {
			name: "single_nested",
			attr: Attribute{
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "single_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-list": {
			name: "list_nested",
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-list-min": {
			name: "list_nested",
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
					MinItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
				},
			},
		},
		"nested-attr-list-max": {
			name: "list_nested",
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
					MaxItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MaxItems: 1,
				},
			},
		},
		"nested-attr-list-minmax": {
			name: "list_nested",
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
					MinItems: 1,
					MaxItems: 10,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
					MaxItems: 10,
				},
			},
		},
		"nested-attr-set": {
			name: "set_nested",
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-set-min": {
			name: "set_nested",
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
					MinItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
				},
			},
		},
		"nested-attr-set-max": {
			name: "set_nested",
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
					MaxItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MaxItems: 1,
				},
			},
		},
		"nested-attr-set-minmax": {
			name: "set_nested",
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
					MinItems: 1,
					MaxItems: 10,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
					MaxItems: 10,
				},
			},
		},
		"attr-and-nested-attr-set": {
			name: "whoops",
			attr: Attribute{
				Type: types.StringType,
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"testing": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "can't have both Attributes and Type set",
		},
		"attr-and-nested-attr-unset": {
			name: "whoops",
			attr: Attribute{
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
		"attr-and-nested-attr-empty": {
			name: "whoops",
			attr: Attribute{
				Optional:   true,
				Attributes: SingleNestedAttributes(map[string]Attribute{}),
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.attr.tfprotov6SchemaAttribute(context.Background(), tc.name, tc.path)
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

func TestAttributeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateAttributeRequest
		resp ValidateAttributeResponse
	}{
		"no-attributes-or-type": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity:  tfprotov6.DiagnosticSeverityError,
						Summary:   "Invalid Attribute Definition",
						Detail:    "Attribute must define either Attributes or Type. This is always a problem with the provider and should be reported to the provider developer.",
						Attribute: tftypes.NewAttributePath().WithAttributeName("test"),
					},
				},
			},
		},
		"both-attributes-and-type": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
									},
								}),
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity:  tfprotov6.DiagnosticSeverityError,
						Summary:   "Invalid Attribute Definition",
						Detail:    "Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
						Attribute: tftypes.NewAttributePath().WithAttributeName("test"),
					},
				},
			},
		},
		"config-error": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nottest": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"nottest": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity:  tfprotov6.DiagnosticSeverityError,
						Summary:   "Configuration Read Error",
						Detail:    "An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\nAttributeName(\"test\") still remains in the path: step cannot be applied to this value",
						Attribute: tftypes.NewAttributePath().WithAttributeName("test"),
					},
				},
			},
		},
		"no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"warnings": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
									testWarningAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testWarningDiagnostic,
					testWarningDiagnostic,
				},
			},
		},
		"errors": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
									testErrorAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testErrorDiagnostic,
					testErrorDiagnostic,
				},
			},
		},
		"type-with-validate-error": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     testtypes.StringTypeWithValidateError{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testtypes.TestErrorDiagnostic,
				},
			},
		},
		"type-with-validate-warning": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     testtypes.StringTypeWithValidateWarning{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testtypes.TestWarningDiagnostic,
				},
			},
		},
		"nested-attr-list-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: ListNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, ListNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-list-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: ListNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, ListNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testErrorDiagnostic,
				},
			},
		},
		"nested-attr-map-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
									AttributeType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{
									AttributeType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: MapNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, MapNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-map-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
									AttributeType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{
									AttributeType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: MapNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, MapNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testErrorDiagnostic,
				},
			},
		},
		"nested-attr-single-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-single-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					testErrorDiagnostic,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateAttributeResponse
			attribute, err := tc.req.Config.Schema.AttributeAtPath(tc.req.AttributePath)

			if err != nil {
				t.Fatalf("Unexpected error getting Attribute: %s", err)
			}

			attribute.validate(context.Background(), tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
