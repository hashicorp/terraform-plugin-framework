// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Request information for the provider logic to update a resource identity
// from a prior resource identity version to the current identity version.
type UpgradeResourceIdentityRequest struct {
	// TypeName is the type of resource that Terraform needs to upgrade the
	// identity state for.
	TypeName string

	// Version is the version of the identity state the resource currently has.
	Version int64

	// RawIdentity is the identity state as Terraform sees it right now in JSON. See the
	// documentation for `RawIdentity` for information on how to work with the
	// data it contains.
	RawState *tfprotov6.RawState

	// Previous identity of the resource if the wrapping IdentityUpgrader
	// type PriorSchema field was present. When available, this allows for
	// easier data handling such as calling Get() or GetAttribute().
	Identity *tfsdk.ResourceIdentity
}

// Response information for the provider logic to update a resource identity
// from a prior resource identity version to the current identity version.
type UpgradeResourceIdentityResponse struct {
	Identity *tfsdk.ResourceIdentity

	// Diagnostics report errors or warnings related to retrieving the resource
	// identity schema. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics
}
