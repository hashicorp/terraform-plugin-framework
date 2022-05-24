package fwserver

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PlanResourceChangeRequest is the framework server request for the
// PlanResourceChange RPC.
type PlanResourceChangeRequest struct {
	Config           *tfsdk.Config
	PriorPrivate     []byte
	PriorState       *tfsdk.State
	ProposedNewState *tfsdk.Plan
	ProviderMeta     *tfsdk.Config
	ResourceSchema   tfsdk.Schema
	ResourceType     tfsdk.ResourceType
}

// PlanResourceChangeResponse is the framework server response for the
// PlanResourceChange RPC.
type PlanResourceChangeResponse struct {
	Diagnostics    diag.Diagnostics
	PlannedPrivate []byte
	PlannedState   *tfsdk.State

	// TODO: Replace with framework defined type
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/81
	RequiresReplace []*tftypes.AttributePath
}

// PlanResourceChange implements the framework server PlanResourceChange RPC.
func (s *Server) PlanResourceChange(ctx context.Context, req *PlanResourceChangeRequest, resp *PlanResourceChangeResponse) {
	if req == nil {
		return
	}

	// Always instantiate new Resource instances.
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := req.ResourceType.NewResource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	nullTfValue := tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil)

	// Prevent potential panics by ensuring incoming Config/Plan/State are null
	// instead of nil.
	if req.Config == nil {
		req.Config = &tfsdk.Config{
			Raw:    nullTfValue,
			Schema: req.ResourceSchema,
		}
	}

	if req.ProposedNewState == nil {
		req.ProposedNewState = &tfsdk.Plan{
			Raw:    nullTfValue,
			Schema: req.ResourceSchema,
		}
	}

	if req.PriorState == nil {
		req.PriorState = &tfsdk.State{
			Raw:    nullTfValue,
			Schema: req.ResourceSchema,
		}
	}

	resp.PlannedState = planToState(*req.ProposedNewState)

	// Execute any AttributePlanModifiers.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers are run again after that marking to allow setting values
	// and preventing extraneous plan differences.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// Execute any resource-level ModifyPlan method.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers be run again after that marking to allow setting values and
	// preventing extraneous plan differences.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// After ensuring there are proposed changes, mark any computed attributes
	// that are null in the config as unknown in the plan, so providers have
	// the choice to update them.
	//
	// Later attribute and resource plan modifier passes can override the
	// unknown with a known value using any plan modifiers.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !resp.PlannedState.Raw.IsNull() && !resp.PlannedState.Raw.Equal(req.PriorState.Raw) {
		logging.FrameworkTrace(ctx, "Marking Computed null Config values as unknown in Plan")

		modifiedPlan, err := tftypes.Transform(resp.PlannedState.Raw, MarkComputedNilsAsUnknown(ctx, req.Config.Raw, req.ResourceSchema))

		if err != nil {
			resp.Diagnostics.AddError(
				"Error modifying plan",
				"There was an unexpected error updating the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if !resp.PlannedState.Raw.Equal(modifiedPlan) {
			logging.FrameworkTrace(ctx, "At least one Computed null Config value was changed to unknown")
		}

		resp.PlannedState.Raw = modifiedPlan
	}

	// Execute any AttributePlanModifiers again. This allows overwriting
	// any unknown values.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !resp.PlannedState.Raw.IsNull() {
		modifySchemaPlanReq := ModifySchemaPlanRequest{
			Config: *req.Config,
			Plan:   stateToPlan(*resp.PlannedState),
			State:  *req.PriorState,
		}

		if req.ProviderMeta != nil {
			modifySchemaPlanReq.ProviderMeta = *req.ProviderMeta
		}

		modifySchemaPlanResp := ModifySchemaPlanResponse{
			Diagnostics: resp.Diagnostics,
			Plan:        modifySchemaPlanReq.Plan,
		}

		SchemaModifyPlan(ctx, req.ResourceSchema, modifySchemaPlanReq, &modifySchemaPlanResp)

		resp.Diagnostics = modifySchemaPlanResp.Diagnostics
		resp.PlannedState = planToState(modifySchemaPlanResp.Plan)
		resp.RequiresReplace = append(resp.RequiresReplace, modifySchemaPlanResp.RequiresReplace...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Execute any resource-level ModifyPlan method again. This allows
	// overwriting any unknown values.
	//
	// We do this regardless of whether the plan is null or not, because we
	// want resources to be able to return diagnostics when planning to
	// delete resources, e.g. to inform practitioners that the resource
	// _can't_ be deleted in the API and will just be removed from
	// Terraform's state
	if resource, ok := resource.(tfsdk.ResourceWithModifyPlan); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithModifyPlan")

		modifyPlanReq := tfsdk.ModifyResourcePlanRequest{
			Config: *req.Config,
			Plan:   stateToPlan(*resp.PlannedState),
			State:  *req.PriorState,
		}

		if req.ProviderMeta != nil {
			modifyPlanReq.ProviderMeta = *req.ProviderMeta
		}

		modifyPlanResp := tfsdk.ModifyResourcePlanResponse{
			Diagnostics:     resp.Diagnostics,
			Plan:            modifyPlanReq.Plan,
			RequiresReplace: []*tftypes.AttributePath{},
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Resource ModifyPlan")
		resource.ModifyPlan(ctx, modifyPlanReq, &modifyPlanResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource ModifyPlan")

		resp.Diagnostics = modifyPlanResp.Diagnostics
		resp.PlannedState = planToState(modifyPlanResp.Plan)
		resp.RequiresReplace = append(resp.RequiresReplace, modifyPlanResp.RequiresReplace...)
	}

	// Ensure deterministic RequiresReplace by sorting and deduplicating
	resp.RequiresReplace = NormaliseRequiresReplace(ctx, resp.RequiresReplace)
}

func MarkComputedNilsAsUnknown(ctx context.Context, config tftypes.Value, resourceSchema tfsdk.Schema) func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error) {
	return func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		ctx = logging.FrameworkWithAttributePath(ctx, path.String())

		// we are only modifying attributes, not the entire resource
		if len(path.Steps()) < 1 {
			return val, nil
		}

		configVal, _, err := tftypes.WalkAttributePath(config, path)

		if err != tftypes.ErrInvalidStep && err != nil {
			logging.FrameworkError(ctx, "error walking attribute path")
			return val, err
		} else if err != tftypes.ErrInvalidStep && !configVal.(tftypes.Value).IsNull() {
			logging.FrameworkTrace(ctx, "attribute not null in config, not marking unknown")
			return val, nil
		}

		attribute, err := resourceSchema.AttributeAtPath(path)

		if err != nil {
			if errors.Is(err, tfsdk.ErrPathInsideAtomicAttribute) {
				// ignore attributes/elements inside schema.Attributes, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is a non-schema attribute, not marking unknown")
				return val, nil
			}

			logging.FrameworkError(ctx, "couldn't find attribute in resource schema")

			return tftypes.Value{}, fmt.Errorf("couldn't find attribute in resource schema: %w", err)
		}
		if !attribute.Computed {
			logging.FrameworkTrace(ctx, "attribute is not computed in schema, not marking unknown")

			return val, nil
		}

		logging.FrameworkDebug(ctx, "marking computed attribute that is null in the config as unknown")

		return tftypes.NewValue(val.Type(), tftypes.UnknownValue), nil
	}
}

// NormaliseRequiresReplace sorts and deduplicates the slice of AttributePaths
// used in the RequiresReplace response field.
// Sorting is lexical based on the string representation of each AttributePath.
func NormaliseRequiresReplace(ctx context.Context, rs []*tftypes.AttributePath) []*tftypes.AttributePath {
	if len(rs) < 2 {
		return rs
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].String() < rs[j].String()
	})

	ret := make([]*tftypes.AttributePath, len(rs))
	ret[0] = rs[0]

	// deduplicate
	j := 1

	for i := 1; i < len(rs); i++ {
		if rs[i].Equal(ret[j-1]) {
			logging.FrameworkDebug(ctx, "attribute found multiple times in RequiresReplace, removing duplicate", map[string]interface{}{logging.KeyAttributePath: rs[i]})
			continue
		}
		ret[j] = rs[i]
		j++
	}

	return ret[:j]
}

// planToState returns a *tfsdk.State with a copied value from a tfsdk.Plan.
func planToState(plan tfsdk.Plan) *tfsdk.State {
	return &tfsdk.State{
		Raw:    plan.Raw.Copy(),
		Schema: plan.Schema,
	}
}

// stateToPlan returns a tfsdk.Plan with a copied value from a tfsdk.State.
func stateToPlan(state tfsdk.State) tfsdk.Plan {
	return tfsdk.Plan{
		Raw:    state.Raw.Copy(),
		Schema: state.Schema,
	}
}
