// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Float64ParameterWithValidate extends the basetypes.Float64Valuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type Float64ParameterWithValidate interface {
	basetypes.Float64Valuable

	ValidateableParameter
}
