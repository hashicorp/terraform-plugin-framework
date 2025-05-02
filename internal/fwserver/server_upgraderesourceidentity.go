// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// UpgradeResourceIdentityRequest is the framework server request for the
// UpgradeResourceIdentity RPC.
type UpgradeResourceIdentityRequest struct {
	Resource       resource.Resource
	IdentitySchema fwschema.Schema
	// TypeName is the type of resource that Terraform needs to upgrade the
	// identity state for.
	TypeName string

	// Version is the version of the identity state the resource currently has.
	Version int64

	// Using the tfprotov6 type here was a pragmatic effort decision around when
	// the framework introduced compatibility promises. This type was chosen as
	// it was readily available and trivial to convert between tfprotov5.
	//
	// Using a terraform-plugin-go type is not ideal for the framework as almost
	// all terraform-plugin-go types have framework abstractions, but if there
	// is ever a time where it makes sense to re-evaluate this decision, such as
	// a major version bump, it could be changed then.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/340
	RawState *tfprotov6.RawState
}

// UpgradeResourceIdentityResponse is the framework server response for the
// UpgradeResourceIdentity RPC.
type UpgradeResourceIdentityResponse struct {
	UpgradedIdentity *tfsdk.ResourceIdentity
	Diagnostics      diag.Diagnostics
}

// UpgradeResourceIdentity implements the framework server UpgradeResourceIdentity RPC.
func (s *Server) UpgradeResourceIdentity(ctx context.Context, req *UpgradeResourceIdentityRequest, resp *UpgradeResourceIdentityResponse) {
	if req == nil {
		return
	}

	// No UpgradedIdentity to return. This could return an error diagnostic about
	// the odd scenario, but seems best to allow Terraform CLI to handle the
	// situation itself in case it might be expected behavior.
	if req.RawState == nil {
		return
	}

	// Define options to be used when unmarshalling raw state.
	// IgnoreUndefinedAttributes will silently skip over fields in the JSON
	// that do not have a matching entry in the schema.
	unmarshalOpts := tfprotov6.UnmarshalOpts{
		ValueFromJSONOpts: tftypes.ValueFromJSONOpts{
			IgnoreUndefinedAttributes: true,
		},
	}

	// TODO: Maybe throw error if the schemas are the same, it is a bug in core

	// Terraform CLI can call UpgradeResourceIdentity even if the stored Identity
	// version matches the current schema. Presumably this is to account for
	// the previous terraform-plugin-sdk implementation, which handled some
	// Identity fixups on behalf of Terraform CLI. When this happens, we do not
	// want to return errors for a missing ResourceWithUpgradeIdentity
	// implementation or an undefined version within an existing
	// ResourceWithUpgradeIdentity implementation as that would be confusing
	// detail for provider developers. Instead, the framework will attempt to
	// roundtrip the prior RawState to a Identity matching the current Schema.
	//
	// TODO: To prevent provider developers from accidentally implementing
	// ResourceWithUpgradeIdentity with a version matching the current schema
	// version which would never get called, the framework can introduce a
	// unit test helper.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/113
	//
	// UnmarshalWithOpts allows optionally ignoring instances in which elements being
	// do not have a corresponding attribute within the schema.
	/*	if req.Version == req.IdentitySchema.GetVersion() {
		logging.FrameworkTrace(ctx, "UpgradeResourceIdentity request version matches current Schema version, using framework defined passthrough implementation")

		identitySchemaType := req.IdentitySchema.Type().TerraformType(ctx)

		rawStateValue, err := req.RawState.UnmarshalWithOpts(identitySchemaType, unmarshalOpts)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
				"There was an error reading the saved resource Identity using the current resource schema.\n\n"+
					"If this resource Identity was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. "+
					"If you manually modified the resource Identity, you will need to manually modify it to match the current resource schema. "+
					"Otherwise, please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		resp.UpgradedIdentity = &tfsdk.ResourceIdentity{
			Schema: req.IdentitySchema,
			Raw:    rawStateValue,
		}

		return
	}*/

	if resourceWithConfigure, ok := req.Resource.(resource.ResourceWithConfigure); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithConfigure")

		configureReq := resource.ConfigureRequest{
			ProviderData: s.ResourceConfigureData,
		}
		configureResp := resource.ConfigureResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined Resource Configure")
		resourceWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined Resource Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	resourceWithUpgradeIdentity, ok := req.Resource.(resource.ResourceWithUpgradeIdentity)

	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Upgrade Resource Identity",
			"This resource was implemented without an UpgradeIdentity() method, "+
				fmt.Sprintf("however Terraform was expecting an implementation for version %d upgrade.\n\n", req.Version)+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	logging.FrameworkTrace(ctx, "Resource implements ResourceWithUpgradeIdentity")

	logging.FrameworkTrace(ctx, "Calling provider defined Resource UpgradeIdentity")
	resourceIdentityUpgraders := resourceWithUpgradeIdentity.UpgradeResourceIdentity(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined Resource UpgradeIdentity")

	// Panic prevention
	if resourceIdentityUpgraders == nil {
		resourceIdentityUpgraders = make(map[int64]resource.IdentityUpgrader, 0)
	}

	resourceIdentityUpgrader, ok := resourceIdentityUpgraders[req.Version]

	if !ok {
		resp.Diagnostics.AddError(
			"Unable to Upgrade Resource Identity",
			"This resource was implemented with an UpgradeIdentity() method, "+
				fmt.Sprintf("however Terraform was expecting an implementation for version %d upgrade.\n\n", req.Version)+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	upgradeResourceIdentityRequest := resource.UpgradeResourceIdentityRequest{
		RawState: req.RawState,
	}

	if resourceIdentityUpgrader.PriorSchema != nil {
		logging.FrameworkTrace(ctx, "Initializing populated UpgradeResourceIdentityRequest Identity from provider defined prior schema and request RawState")

		priorSchemaType := resourceIdentityUpgrader.PriorSchema.Type().TerraformType(ctx)

		_, err := req.RawState.UnmarshalWithOpts(priorSchemaType, unmarshalOpts)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Previously Saved Identity for UpgradeResourceIdentity",
				fmt.Sprintf("There was an error reading the saved resource Identity using the prior resource schema defined for version %d upgrade.\n\n", req.Version)+
					"Please report this to the provider developer:\n\n"+err.Error(),
			)
			return
		}

		upgradeResourceIdentityRequest.RawState = req.RawState

	}

	upgradeResourceIdentityResponse := resource.UpgradeResourceIdentityResponse{
		UpgradedIdentity: &tfsdk.ResourceIdentity{
			Schema: req.IdentitySchema,
			// Raw is intentionally not set.
		},
	}

	// To simplify provider logic, this could perform a best effort attempt
	// to populate the response Identity by looping through all Attribute/Block
	// by calling the equivalent of SetAttribute(GetAttribute()) and skipping
	// any errors.

	logging.FrameworkTrace(ctx, "Calling provider defined IdentityUpgrader")
	resourceIdentityUpgrader.IdentityUpgrader(ctx, upgradeResourceIdentityRequest, &upgradeResourceIdentityResponse)
	logging.FrameworkTrace(ctx, "Called provider defined IdentityUpgrader")

	resp.Diagnostics.Append(upgradeResourceIdentityResponse.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	if upgradeResourceIdentityResponse.UpgradedIdentity.Raw.Type() == nil || upgradeResourceIdentityResponse.UpgradedIdentity.Raw.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Upgraded Resource Identity",
			fmt.Sprintf("After attempting a resource Identity upgrade to version %d, the provider did not return any Identity data. ", req.Version)+
				"Preventing the unexpected loss of resource Identity data. "+
				"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
		)
		return
	}

	if upgradeResourceIdentityResponse.UpgradedIdentity != nil {
		logging.FrameworkTrace(ctx, "UpgradeResourceIdentityResponse Raw State set, overriding State")

		resp.UpgradedIdentity = &tfsdk.ResourceIdentity{
			Schema: req.IdentitySchema,
			Raw:    upgradeResourceIdentityResponse.UpgradedIdentity.Raw,
		}

		return
	}

	resp.UpgradedIdentity = upgradeResourceIdentityResponse.UpgradedIdentity
}
