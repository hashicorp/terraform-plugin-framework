// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ValueSemanticEqualityFloat32 performs float32 type semantic equality.
func ValueSemanticEqualityFloat32(ctx context.Context, req ValueSemanticEqualityRequest, resp *ValueSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.Float32ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.Float32ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.Float32SemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}
