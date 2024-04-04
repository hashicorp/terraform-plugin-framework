// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xfwfunction

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

// Parameter returns the Parameter for a given argument position. This may be
// from the Parameters field or, if defined, the VariadicParameter field. An
// error diagnostic is raised if the position is outside the expected arguments.
func Parameter(ctx context.Context, d function.Definition, position int) (function.Parameter, *function.FuncError) {
	if d.VariadicParameter != nil && position >= len(d.Parameters) {
		return d.VariadicParameter, nil
	}

	pos := int64(position)

	if len(d.Parameters) == 0 {
		return nil, function.NewArgumentFuncError(
			pos,
			"Invalid Parameter Position for Definition: "+
				"When determining the parameter for the given argument position, an invalid value was given. "+
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
				"Function does not implement parameters.\n"+
				fmt.Sprintf("Given position: %d", position),
		)
	}

	if position >= len(d.Parameters) {
		return nil, function.NewArgumentFuncError(
			pos,
			"Invalid Parameter Position for Definition: "+
				"When determining the parameter for the given argument position, an invalid value was given. "+
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Max argument position: %d\n", len(d.Parameters)-1)+
				fmt.Sprintf("Given position: %d", position),
		)
	}

	return d.Parameters[position], nil
}
