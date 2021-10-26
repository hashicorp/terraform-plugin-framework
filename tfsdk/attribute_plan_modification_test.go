package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPreserveStateModifier(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state    attr.Value
		plan     attr.Value
		config   attr.Value
		expected attr.Value
	}

	tests := map[string]testCase{
		"nil-state": {
			// this honestly just shouldn't happen, but let's be
			// sure we're not going to panic if it does
			state:    nil,
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Unknown: true},
		},
		"nil-plan": {
			// this honestly just shouldn't happen, but let's be
			// sure we're not going to panic if it does
			state:    types.String{Null: true},
			plan:     nil,
			config:   types.String{Null: true},
			expected: nil,
		},
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			state:    types.String{Null: true},
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Unknown: true},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			state:    types.String{Value: "foo"},
			plan:     types.String{Value: "bar"},
			config:   types.String{Null: true},
			expected: types.String{Value: "bar"},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			state:    types.String{Value: "foo"},
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Value: "foo"},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			state:    types.String{Value: "foo"},
			plan:     types.String{Unknown: true},
			config:   types.String{Unknown: true},
			expected: types.String{Unknown: true},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schema := Schema{
				Attributes: map[string]Attribute{
					"a": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				},
			}

			var configRaw, planRaw, stateRaw interface{}
			if tc.config != nil {
				val, err := tc.config.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				configRaw = val
			}
			if tc.state != nil {
				val, err := tc.state.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				stateRaw = val
			}
			if tc.plan != nil {
				val, err := tc.plan.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				planRaw = val
			}
			configVal := tftypes.NewValue(tftypes.String, configRaw)
			stateVal := tftypes.NewValue(tftypes.String, stateRaw)
			planVal := tftypes.NewValue(tftypes.String, planRaw)

			req := ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath(),
				Config: Config{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": configVal,
					}),
				},
				State: State{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": stateVal,
					}),
				},
				Plan: Plan{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": planVal,
					}),
				},
				AttributeConfig: tc.config,
				AttributeState:  tc.state,
				AttributePlan:   tc.plan,
				ProviderMeta:    Config{},
			}
			resp := &ModifyAttributePlanResponse{
				AttributePlan: req.AttributePlan,
			}
			modifier := PreserveState()

			modifier.Modify(context.Background(), req, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Unexpected diagnostics: %s", resp.Diagnostics)
			}
			if diff := cmp.Diff(tc.expected, resp.AttributePlan); diff != "" {
				t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
			}
		})
	}
}
