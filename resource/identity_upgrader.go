// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Implementation handler for a UpgradeState operation.
//
// This is used to encapsulate all upgrade logic from a prior state to the
// current schema version when a Resource implements the
// ResourceWithUpgradeState interface.
type IdentityUpgrader struct {
	// Schema information for the prior state version. While not required,
	// setting this will populate the UpgradeStateRequest type State
	// field similar to other Resource data types. This allows for easier data
	// handling such as calling Get() or GetAttribute().
	//
	// If not set, prior state data is available in the
	// UpgradeResourceStateRequest type RawState field.
	PriorSchema *identityschema.Schema

	// Version is the version schema that this Upgrader will handle, converting
	// it to Version+1.
	Version int64

	// Type describes the schema that this function can upgrade. Type is
	// required to decode the schema if the state was stored in a legacy
	// flatmap format.
	Type tftypes.Type

	// Provider defined logic for upgrading a resource state from the prior
	// state version to the current schema version.
	//
	// The context.Context parameter contains framework-defined loggers and
	// supports request cancellation.
	//
	// The UpgradeStateRequest parameter contains the prior state data.
	// If PriorSchema was set, the State field will be available. Otherwise,
	// the RawState must be used.
	//
	// The UpgradeStateResponse parameter should contain the upgraded
	// state data and can be used to signal any logic warnings or errors.
	IdentityUpgrader func(context.Context, UpgradeResourceIdentityRequest, *UpgradeResourceIdentityResponse)
}
