package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (rt testServeResourceTypeAttributePlanModifiers) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version: 1,
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Required: true,
				Type:     types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					// For the purposes of testing, these plan modifiers behave
					// differently for certain values of the attribute.
					// By default, they do nothing.
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
				},
			},
			"size": {
				Required: true,
				Type:     types.NumberType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplaceIf(func(ctx context.Context, state, config attr.Value, path *tftypes.AttributePath) (bool, diag.Diagnostics) {
					if state == nil && config == nil {
						return false, nil
					}
					if (state == nil && config != nil) || (state != nil && config == nil) {
						return true, nil
					}
					var stateVal, configVal types.Number
					diags := tfsdk.ValueAs(ctx, state, &stateVal)
					if diags.HasError() {
						return false, diags
					}
					diags.Append(tfsdk.ValueAs(ctx, config, &configVal)...)
					if diags.HasError() {
						return false, diags
					}

					if !stateVal.Unknown && !stateVal.Null && !configVal.Unknown && !configVal.Null {
						if configVal.Value.Cmp(stateVal.Value) > 0 {
							return true, diags
						}
					}
					return false, diags
				}, "If the new size is greater than the old size, Terraform will destroy and recreate the resource", "If the new size is greater than the old size, Terraform will destroy and recreate the resource"),
				}},
			"scratch_disk": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Required: true,
						Type:     types.StringType,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							planmodifiers.TestAttrPlanValueModifierTwo{},
						},
					},
					"interface": {
						Required:      true,
						Type:          types.StringType,
						PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
					},
					"filesystem": {
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"size": {
								Optional: true,
								Type:     types.NumberType,
							},
							"format": {
								Optional:      true,
								Type:          types.StringType,
								PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
							},
						}),
					},
				}),
			},
			"region": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{planmodifiers.TestAttrDefaultValueModifier{}},
			},
			"computed_string_no_modifiers": {
				Computed: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (rt testServeResourceTypeAttributePlanModifiers) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeAttributePlanModifiers{
		provider: provider,
	}, nil
}

var testServeResourceTypeAttributePlanModifiersType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"computed_string_no_modifiers": tftypes.String,
		"name":                         tftypes.String,
		"size":                         tftypes.Number,
		"scratch_disk": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"id":        tftypes.String,
				"interface": tftypes.String,
				"filesystem": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					},
				},
			},
		},
		"region": tftypes.String,
	},
}

type testServeAttributePlanModifiers struct {
	provider *testServeProvider
}

type testServeResourceTypeAttributePlanModifiers struct{}

func (r testServeAttributePlanModifiers) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeAttributePlanModifiers) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeAttributePlanModifiers) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeAttributePlanModifiers) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
