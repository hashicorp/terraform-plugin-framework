// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

// SchemaRequest represents a request for the Action to return its schema.
// An instance of this request struct is supplied as an argument to the
// Action type Schema method.
type SchemaRequest struct{}

// SchemaResponse represents a response to a SchemaRequest. An instance of this
// response struct is supplied as an argument to the Action type Schema
// method.
type SchemaResponse struct {
	// TODO:Actions: This will eventually be replaced by an interface defined in
	// an "actions/schema" package. Schema implementations that will fulfill this
	// interface will be unlinked, linked, or lifecycle. (also defined in the "actions/schema" package)
	Schema fwschema.Schema

	// Diagnostics report errors or warnings related to retrieving the action schema.
	// An empty slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}
