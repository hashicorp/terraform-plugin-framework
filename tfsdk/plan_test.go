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

func TestPlanGet(t *testing.T) {
	t.Parallel()

	type testPlanGetData struct {
		Name types.String `tfsdk:"name"`
	}

	type testCase struct {
		plan          Plan
		expected      testPlanGetData
		expectedDiags []*tfprotov6.Diagnostic
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetData{
				Name: types.String{Value: "namevalue"},
			},
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetData{
				Name: types.String{Value: ""},
			},
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestErrorDiagnostic},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetData{
				Name: types.String{Value: "namevalue"},
			},
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestWarningDiagnostic},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testPlanGetData

			diags := tc.plan.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		expected      attr.Value
		expectedDiags []*tfprotov6.Diagnostic
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			expected: types.String{Value: "namevalue"},
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			expected:      nil,
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestErrorDiagnostic},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			expected:      types.String{Value: "namevalue"},
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestWarningDiagnostic},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.plan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		val           interface{}
		expected      tftypes.Value
		expectedDiags []*tfprotov6.Diagnostic
	}

	testCases := map[string]testCase{
		"write": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "oldvalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"overwrite": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected:      tftypes.Value{},
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestErrorDiagnostic},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestWarningDiagnostic},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		path          *tftypes.AttributePath
		val           interface{}
		expected      tftypes.Value
		expectedDiags []*tfprotov6.Diagnostic
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{testtypes.TestErrorDiagnostic},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: []*tfprotov6.Diagnostic{
				testtypes.TestWarningDiagnostic,
				// TODO: Consider duplicate diagnostic consolidation functionality with diagnostic abstraction
				// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/24
				testtypes.TestWarningDiagnostic,
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
