package tfsdk

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type typeWithPlanModifier struct {
	modifyPlan func(ctx context.Context, state attr.Value, plan attr.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics)
}

func (t typeWithPlanModifier) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t typeWithPlanModifier) ValueFromTerraform(_ context.Context, val tftypes.Value) (attr.Value, error) {
	ret := testtypes.String{CreatedBy: t}
	if val.IsNull() {
		ret.String = types.String{Null: true}
		return ret, nil
	}
	if !val.IsKnown() {
		ret.String = types.String{Unknown: true}
		return ret, nil
	}
	var v string
	err := val.As(&v)
	if err != nil {
		return nil, err
	}
	ret.String = types.String{Value: v}
	return ret, nil
}

func (t typeWithPlanModifier) Equal(o attr.Type) bool {
	_, ok := o.(typeWithPlanModifier)
	if !ok {
		return false
	}
	return true
}

func (t typeWithPlanModifier) String() string {
	return "tfsdk.typeWithPlanModifier"
}

func (t typeWithPlanModifier) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t typeWithPlanModifier) ModifyPlan(ctx context.Context, state attr.Value, plan attr.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	return t.modifyPlan(ctx, state, plan, path)
}

func TestRunTypePlanModifiers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         tftypes.Value
		plan          tftypes.Value
		schema        Schema
		resp          *planResourceChangeResponse
		expectedPlan  tftypes.Value
		expectedDiags diag.Diagnostics
		expectedRR    []*tftypes.AttributePath
		expectedOK    bool
	}

	tests := map[string]testCase{
		"case-insensitive": {
			state: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
			plan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hElLo, WoRlD"),
			}),
			schema: Schema{
				Attributes: map[string]Attribute{
					"input": {
						Type: typeWithPlanModifier{
							modifyPlan: func(ctx context.Context, state attr.Value, plan attr.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
								st := state.(testtypes.String)
								pl := plan.(testtypes.String)
								if strings.ToLower(st.String.Value) == strings.ToLower(pl.String.Value) {
									return state, nil
								}
								return plan, nil
							},
						},
						Required: true,
					},
				},
			},
			resp: &planResourceChangeResponse{},
			expectedPlan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
			expectedDiags: nil,
			expectedRR:    nil,
			expectedOK:    true,
		},
		"preserve-existing": {
			state: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
			plan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hElLo, WoRlD"),
			}),
			schema: Schema{
				Attributes: map[string]Attribute{
					"input": {
						Type: typeWithPlanModifier{
							modifyPlan: func(ctx context.Context, state attr.Value, plan attr.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
								st := state.(testtypes.String)
								pl := plan.(testtypes.String)
								if strings.ToLower(st.String.Value) == strings.ToLower(pl.String.Value) {
									return state, diag.Diagnostics{
										diag.NewWarningDiagnostic(
											"Diff suppressed",
											"We suppressed a diff because the strings were only different in capitalization. Normally you wouldn't warn on this, but work with me here.",
										),
									}
								}
								return plan, nil
							},
						},
						Required: true,
					},
				},
			},
			resp: &planResourceChangeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"Other warning",
						"Deprecated attribute or something",
					),
				},
				RequiresReplace: []*tftypes.AttributePath{
					tftypes.NewAttributePath().WithAttributeName("foo"),
				},
			},
			expectedPlan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					"Other warning",
					"Deprecated attribute or something",
				),
				diag.NewWarningDiagnostic(
					"Diff suppressed",
					"We suppressed a diff because the strings were only different in capitalization. Normally you wouldn't warn on this, but work with me here.",
				),
			},
			expectedRR: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("foo"),
			},
			expectedOK: true,
		},
		"error": {
			state: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
			plan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hElLo, WoRlD"),
			}),
			schema: Schema{
				Attributes: map[string]Attribute{
					"input": {
						Type: typeWithPlanModifier{
							modifyPlan: func(ctx context.Context, state attr.Value, plan attr.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
								// something bad happened
								return plan, diag.Diagnostics{
									diag.NewErrorDiagnostic("Ooops", "something bad happened"),
								}
							},
						},
						Required: true,
					},
				},
			},
			resp: &planResourceChangeResponse{},
			expectedPlan: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"input": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"input": tftypes.NewValue(tftypes.String, "hElLo, WoRlD"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Ooops", "something bad happened"),
			},
			expectedRR: nil,
			expectedOK: false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan, ok := runTypePlanModifiers(context.Background(), tc.state, tc.plan, tc.schema, tc.resp)

			if ok != tc.expectedOK {
				t.Fatalf("expected ok to be %v, got %v", tc.expectedOK, ok)
			}
			if diff := cmp.Diff(tc.resp.Diagnostics, tc.expectedDiags); diff != "" {
				t.Fatalf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(plan, tc.expectedPlan); diff != "" {
				t.Fatalf("Unexpected diff in plan result (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(tc.resp.RequiresReplace, tc.expectedRR); diff != "" {
				t.Fatalf("Unexpected diff in requires replace (+wanted, -got): %s", diff)
			}
		})
	}
}
