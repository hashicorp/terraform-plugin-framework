// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SetParameterWithValidate extends the basetypes.SetValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type SetParameterWithValidate interface {
	basetypes.SetValuable

	ValidateableParameter
}
