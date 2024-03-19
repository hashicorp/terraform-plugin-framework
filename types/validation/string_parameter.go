// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// StringParameterWithValidate extends the basetypes.StringValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type StringParameterWithValidate interface {
	basetypes.StringValuable

	ValidateableParameter
}
