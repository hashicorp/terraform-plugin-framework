package tfsdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ ResourceWithUpgradeState = testServeResourceUpgradeState{}

type testServeResourceTypeUpgradeState struct{}

func (t testServeResourceTypeUpgradeState) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_attribute": {
				Type:     types.StringType,
				Optional: true,
			},
			"required_attribute": {
				Type:     types.StringType,
				Required: true,
			},
		},
		Version: 5,
	}, nil
}

func (t testServeResourceTypeUpgradeState) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceUpgradeState{
		provider: provider,
	}, nil
}

var testServeResourceTypeUpgradeStateSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "optional_attribute",
				Optional: true,
				Type:     tftypes.String,
			},
			{
				Name:     "required_attribute",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
	Version: 5,
}

var (
	testServeResourceTypeUpgradeStateTftypeV0 = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                 tftypes.String,
			"optional_attribute": tftypes.Bool,
			"required_attribute": tftypes.Bool,
		},
	}
	testServeResourceTypeUpgradeStateTftype = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                 tftypes.String,
			"optional_attribute": tftypes.String,
			"required_attribute": tftypes.String,
		},
	}
)

type testServeResourceUpgradeStateDataV1 struct {
	Id                string `json:"id"`
	OptionalAttribute *bool  `json:"optional_attribute,omitempty"`
	RequiredAttribute bool   `json:"required_attribute"`
}

type testServeResourceUpgradeStateDataV2 struct {
	Id                string `tfsdk:"id"`
	OptionalAttribute *bool  `tfsdk:"optional_attribute"`
	RequiredAttribute bool   `tfsdk:"required_attribute"`
}

type testServeResourceUpgradeStateData struct {
	Id                string  `tfsdk:"id"`
	OptionalAttribute *string `tfsdk:"optional_attribute"`
	RequiredAttribute string  `tfsdk:"required_attribute"`
}

type testServeResourceUpgradeState struct {
	provider *testServeProvider
}

func (r testServeResourceUpgradeState) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceUpgradeState) UpgradeState(ctx context.Context) map[int64]ResourceStateUpgrader {
	r.provider.upgradeResourceStateCalledResourceType = "test_upgrade_state"
	return map[int64]ResourceStateUpgrader{
		0: { // Successful state upgrade using RawState.Unmarshal() and DynamicValue
			StateUpgrader: func(ctx context.Context, req UpgradeResourceStateRequest, resp *UpgradeResourceStateResponse) {
				rawStateValue, err := req.RawState.Unmarshal(testServeResourceTypeUpgradeStateTftypeV0)

				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Unmarshal Prior State",
						err.Error(),
					)
					return
				}

				var rawState map[string]tftypes.Value

				if err := rawStateValue.As(&rawState); err != nil {
					resp.Diagnostics.AddError(
						"Unable to Convert Prior State",
						err.Error(),
					)
					return
				}

				var optionalAttributeString *string

				if !rawState["optional_attribute"].IsNull() {
					var optionalAttribute bool

					if err := rawState["optional_attribute"].As(&optionalAttribute); err != nil {
						resp.Diagnostics.AddAttributeError(
							tftypes.NewAttributePath().WithAttributeName("optional_attribute"),
							"Unable to Convert Prior State",
							err.Error(),
						)
						return
					}

					v := fmt.Sprintf("%t", optionalAttribute)
					optionalAttributeString = &v
				}

				var requiredAttribute bool

				if err := rawState["required_attribute"].As(&requiredAttribute); err != nil {
					resp.Diagnostics.AddAttributeError(
						tftypes.NewAttributePath().WithAttributeName("required_attribute"),
						"Unable to Convert Prior State",
						err.Error(),
					)
					return
				}

				dynamicValue, err := tfprotov6.NewDynamicValue(
					testServeResourceTypeUpgradeStateTftype,
					tftypes.NewValue(testServeResourceTypeUpgradeStateTftype, map[string]tftypes.Value{
						"id":                 rawState["id"],
						"optional_attribute": tftypes.NewValue(tftypes.String, optionalAttributeString),
						"required_attribute": tftypes.NewValue(tftypes.String, fmt.Sprintf("%t", requiredAttribute)),
					}),
				)

				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Convert Upgraded State",
						err.Error(),
					)
					return
				}

				resp.DynamicValue = &dynamicValue
			},
		},
		1: { // Successful state upgrade using RawState.JSON and DynamicValue
			StateUpgrader: func(ctx context.Context, req UpgradeResourceStateRequest, resp *UpgradeResourceStateResponse) {
				var rawState testServeResourceUpgradeStateDataV1

				if err := json.Unmarshal(req.RawState.JSON, &rawState); err != nil {
					resp.Diagnostics.AddError(
						"Unable to Unmarshal Prior State",
						err.Error(),
					)
					return
				}

				var optionalAttribute *string

				if rawState.OptionalAttribute != nil {
					v := fmt.Sprintf("%t", *rawState.OptionalAttribute)
					optionalAttribute = &v
				}

				dynamicValue, err := tfprotov6.NewDynamicValue(
					testServeResourceTypeUpgradeStateTftype,
					tftypes.NewValue(testServeResourceTypeUpgradeStateTftype, map[string]tftypes.Value{
						"id":                 tftypes.NewValue(tftypes.String, rawState.Id),
						"optional_attribute": tftypes.NewValue(tftypes.String, optionalAttribute),
						"required_attribute": tftypes.NewValue(tftypes.String, fmt.Sprintf("%t", rawState.RequiredAttribute)),
					}),
				)

				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Create Upgraded State",
						err.Error(),
					)
					return
				}

				resp.DynamicValue = &dynamicValue
			},
		},
		2: { // Successful state upgrade with PriorSchema and State
			PriorSchema: &Schema{
				Attributes: map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
					},
					"optional_attribute": {
						Type:     types.BoolType,
						Optional: true,
					},
					"required_attribute": {
						Type:     types.BoolType,
						Required: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req UpgradeResourceStateRequest, resp *UpgradeResourceStateResponse) {
				var priorStateData testServeResourceUpgradeStateDataV2

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

				if resp.Diagnostics.HasError() {
					return
				}

				upgradedStateData := testServeResourceUpgradeStateData{
					Id:                priorStateData.Id,
					RequiredAttribute: fmt.Sprintf("%t", priorStateData.RequiredAttribute),
				}

				if priorStateData.OptionalAttribute != nil {
					v := fmt.Sprintf("%t", *priorStateData.OptionalAttribute)
					upgradedStateData.OptionalAttribute = &v
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
		3: { // Incorrect PriorSchema
			PriorSchema: &Schema{
				Attributes: map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
					},
					"optional_attribute": {
						Type:     types.Int64Type, // Purposefully incorrect
						Optional: true,
					},
					"required_attribute": {
						Type:     types.Int64Type, // Purposefully incorrect
						Required: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req UpgradeResourceStateRequest, resp *UpgradeResourceStateResponse) {
				// Expect error before reaching this logic.
			},
		},
		4: { // Missing upgraded resource data
			StateUpgrader: func(ctx context.Context, req UpgradeResourceStateRequest, resp *UpgradeResourceStateResponse) {
				// Purposfully not setting resp.DynamicValue or resp.State
			},
		},
	}
}
