// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type SchemaRequest struct{}

type SchemaResponse struct {
	LinkedResources LinkedResources

	// Schema is the schema of the action
	Schema schema.Schema

	// Diagnostics report errors or warnings related to validating the data
	// source configuration. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics
}
