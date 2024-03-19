// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Int64ParameterWithValidate extends the basetypes.Int64Valuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type Int64ParameterWithValidate interface {
	basetypes.Int64Valuable

	ValidateableParameter
}
