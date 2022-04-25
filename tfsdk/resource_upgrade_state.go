package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Optional interface on top of Resource that enables provider control over
// the UpgradeResourceState RPC. This RPC is automatically called by Terraform
// when the current Schema type Version field is greater than the stored state.
// Terraform does not store previous Schema information, so any breaking
// changes to state data types must be handled by providers.
//
// Terraform CLI can execute the UpgradeResourceState RPC even when the prior
// state version matches the current schema version. The framework will
// automatically intercept this request and attempt to respond with the
// existing state. In this situation the framework will not execute any
// provider defined logic, so declaring it for this version is extraneous.
type ResourceWithUpgradeState interface {
	// A mapping of prior state version to current schema version state upgrade
	// implementations. Only the specified state upgrader for the prior state
	// version is called, rather than each version in between, so it must
	// encapsulate all logic to convert the prior state to the current schema
	// version.
	//
	// Version keys begin at 0, which is the default schema version when
	// undefined. The framework will return an error diagnostic should the
	// requested state version not be implemented.
	UpgradeState(context.Context) map[int64]ResourceStateUpgrader
}

// Implementation handler for a UpgradeResourceState operation.
//
// This is used to encapsulate all upgrade logic from a prior state to the
// current schema version when a Resource implements the
// ResourceWithUpgradeState interface.
type ResourceStateUpgrader struct {
	// Schema information for the prior state version. While not required,
	// setting this will populate the UpgradeResourceStateRequest type State
	// field similar to other Resource data types. This allows for easier data
	// handling such as calling Get() or GetAttribute().
	//
	// If not set, prior state data is available in the
	// UpgradeResourceStateRequest type RawState field.
	PriorSchema *Schema

	// Provider defined logic for upgrading a resource state from the prior
	// state version to the current schema version.
	//
	// The context.Context parameter contains framework-defined loggers and
	// supports request cancellation.
	//
	// The UpgradeResourceStateRequest parameter contains the prior state data.
	// If PriorSchema was set, the State field will be available. Otherwise,
	// the RawState must be used.
	//
	// The UpgradeResourceStateResponse parameter should contain the upgraded
	// state data and can be used to signal any logic warnings or errors.
	StateUpgrader func(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}

// Request information for the provider logic to update a resource state
// from a prior state version to the current schema version. An instance of
// this is supplied as a parameter to the StateUpgrader function defined in a
// ResourceStateUpgrader, which ultimately comes from a Resource's
// UpgradeState method.
type UpgradeResourceStateRequest struct {
	// Previous state of the resource in JSON (Terraform CLI 0.12 and later)
	// or flatmap format, depending on which version of Terraform CLI last
	// wrote the resource state. This data is always available, regardless
	// whether the wrapping ResourceStateUpgrader type PriorSchema field was
	// present.
	//
	// This is advanced functionality for providers wanting to skip the full
	// redeclaration of older schemas and instead use lower level handlers to
	// transform data. A typical implementation for working with this data will
	// call the Unmarshal() method.
	RawState *tfprotov6.RawState

	// Previous state of the resource if the wrapping ResourceStateUpgrader
	// type PriorSchema field was present. When available, this allows for
	// easier data handling such as calling Get() or GetAttribute().
	State *State
}

// Response information for the provider logic to update a resource state
// from a prior state version to the current schema version. An instance of
// this is supplied as a parameter to the StateUpgrader function defined in a
// ResourceStateUpgrader, which ultimately came from a Resource's
// UpgradeState method.
type UpgradeResourceStateResponse struct {
	// Diagnostics report errors or warnings related to upgrading the resource
	// state. An empty slice indicates a successful operation with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics

	// Upgraded state of the resource, which should match the current schema
	// version. If set, this will override State.
	//
	// This field is intended only for advanced provider functionality, such as
	// skipping the full redeclaration of older schemas or using lower level
	// handlers to transform data. Call tfprotov6.NewDynamicValue() to set this
	// value.
	//
	// All data must be populated to prevent data loss during the upgrade
	// operation. No prior state data is copied automatically.
	DynamicValue *tfprotov6.DynamicValue

	// Upgraded state of the resource, which should match the current schema
	// version. If DynamicValue is set, it will override this value.
	//
	// This field allows for easier data handling such as calling Set() or
	// SetAttribute(). It is generally recommended over working with the lower
	// level types and functionality required for DynamicValue.
	//
	// All data must be populated to prevent data loss during the upgrade
	// operation. No prior state data is copied automatically.
	State State
}
