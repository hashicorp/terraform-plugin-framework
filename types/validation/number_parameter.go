// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// NumberParameterWithValidate extends the basetypes.NumberValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type NumberParameterWithValidate interface {
	basetypes.NumberValuable

	ValidateableParameter
}
