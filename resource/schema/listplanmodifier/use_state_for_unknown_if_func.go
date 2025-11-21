// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForUnknownIfFunc is a conditional function used in the UseStateForUnknownIf
// plan modifier to determine whether the attribute should use the state value for unknown.
type UseStateForUnknownIfFunc func(context.Context, planmodifier.ListRequest, *UseStateForUnknownIfFuncResponse)

// UseStateForUnknownIfFuncResponse is the response type for a UseStateForUnknownIfFunc.
type UseStateForUnknownIfFuncResponse struct {
	// Diagnostics report errors or warnings related to this logic. An empty
	// or unset slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics

	// UseState should be enabled if the state value should be used for the plan value.
	UseState bool
}
