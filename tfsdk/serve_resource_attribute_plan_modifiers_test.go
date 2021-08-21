package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (rt testServeResourceTypeAttributePlanModifiers) GetSchema(_ context.Context) (Schema, []*tfprotov6.Diagnostic) {
	return Schema{
		Version: 1,
		Attributes: map[string]Attribute{
			"name": {
				Required: true,
				Type:     types.StringType,
				// For the purposes of testing, these plan modifiers behave
				// differently for certain values of the attribute.
				// By default, they do nothing.
				PlanModifiers: []AttributePlanModifier{
					testWarningDiagModifier{},
					testErrorDiagModifier{},
					testAttrPlanValueModifierOne{},
					testAttrPlanValueModifierTwo{},
				},
			},
			"size": {
				Required: true,
				Type:     types.NumberType,
				PlanModifiers: []AttributePlanModifier{RequiresReplaceIf(func(ctx context.Context, state, config attr.Value) (bool, error) {
					if state == nil && config == nil {
						return false, nil
					}
					if (state == nil && config != nil) || (state != nil && config == nil) {
						return true, nil
					}
					stateVal := state.(types.Number)
					configVal := config.(types.Number)

					if !stateVal.Unknown && !stateVal.Null && !configVal.Unknown && !configVal.Null {
						if configVal.Value.Cmp(stateVal.Value) > 0 {
							return true, nil
						}
					}
					return false, nil
				}, "If the new size is greater than the old size, Terraform will destroy and recreate the resource", "If the new size is greater than the old size, Terraform will destroy and recreate the resource"),
				}},
			"scratch_disk": {
				Optional: true,
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"id": {
						Required: true,
						Type:     types.StringType,
						PlanModifiers: []AttributePlanModifier{
							testAttrPlanValueModifierTwo{},
						},
					},
					"interface": {
						Required:      true,
						Type:          types.StringType,
						PlanModifiers: []AttributePlanModifier{RequiresReplace()},
					},
				}),
			},
			"region": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []AttributePlanModifier{testAttrDefaultValueModifier{}},
			},
		},
	}, nil
}

func (rt testServeResourceTypeAttributePlanModifiers) NewResource(_ context.Context, p Provider) (Resource, []*tfprotov6.Diagnostic) {
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

var testServeResourceTypeAttributePlanModifiersSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
			{
				Name:     "region",
				Optional: true,
				Type:     tftypes.String,
			},
			{
				Name:     "scratch_disk",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "id",
							Required: true,
							Type:     tftypes.String,
						},
						{
							Name:     "interface",
							Required: true,
							Type:     tftypes.String,
						},
					},
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
				},
			},
			{
				Name:     "size",
				Required: true,
				Type:     tftypes.Number,
			},
		},
	},
}

var testServeResourceTypeAttributePlanModifiersType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
		"size": tftypes.Number,
		"scratch_disk": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"id":        tftypes.String,
				"interface": tftypes.String,
			},
		},
		"region": tftypes.String,
	},
}

type testServeAttributePlanModifiers struct {
	provider *testServeProvider
}

type testServeResourceTypeAttributePlanModifiers struct{}

type testWarningDiagModifier struct{}

func (t testWarningDiagModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTDIAG" {
		resp.Diagnostics = append(resp.Diagnostics,
			&tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityWarning,
				Summary:  "Warning diag",
				Detail:   "This is a warning",
			},
		)
	}
}

func (t testWarningDiagModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testWarningDiagModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testErrorDiagModifier struct{}

func (t testErrorDiagModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTDIAG" {
		resp.Diagnostics = append(resp.Diagnostics,
			&tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Error diag",
				Detail:   "This is an error",
			},
		)
	}
}

func (t testErrorDiagModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testErrorDiagModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testAttrPlanValueModifierOne struct{}

func (t testAttrPlanValueModifierOne) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTATTRONE" {
		resp.AttributePlan = types.String{
			Value: "TESTATTRTWO",
		}
	}
}

func (t testAttrPlanValueModifierOne) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testAttrPlanValueModifierOne) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testAttrPlanValueModifierTwo struct{}

func (t testAttrPlanValueModifierTwo) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrVal, ok := req.AttributePlan.(types.String)
	if !ok {
		return
	}

	if attrVal.Value == "TESTATTRTWO" {
		resp.AttributePlan = types.String{
			Value: "MODIFIED_TWO",
		}
	}
}

func (t testAttrPlanValueModifierTwo) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testAttrPlanValueModifierTwo) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

type testAttrDefaultValueModifier struct{}

func (t testAttrDefaultValueModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	if req.AttributeState == nil && req.AttributeConfig == nil {
		return
	}

	configVal := req.AttributeConfig.(types.String)

	if configVal.Null {
		resp.AttributePlan = types.String{Value: "DEFAULTVALUE"}
	}
}

func (t testAttrDefaultValueModifier) Description(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (t testAttrDefaultValueModifier) MarkdownDescription(ctx context.Context) string {
	return "This plan modifier is for use during testing only"
}

func (r testServeAttributePlanModifiers) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_attribute_plan_modifiers"
	r.provider.applyResourceChangeCalledAction = "create"
	r.provider.createFunc(ctx, req, resp)
}

func (r testServeAttributePlanModifiers) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	r.provider.readResourceCurrentStateValue = req.State.Raw
	r.provider.readResourceCurrentStateSchema = req.State.Schema
	r.provider.readResourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readResourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readResourceCalledResourceType = "test_attribute_plan_modifiers"
	r.provider.readResourceImpl(ctx, req, resp)
}

func (r testServeAttributePlanModifiers) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_attribute_plan_modifiers"
	r.provider.applyResourceChangeCalledAction = "update"
	r.provider.updateFunc(ctx, req, resp)
}

func (r testServeAttributePlanModifiers) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_attribute_plan_modifiers"
	r.provider.applyResourceChangeCalledAction = "delete"
	r.provider.deleteFunc(ctx, req, resp)
}
