// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Float64ValuableWithValidateableAttribute extends the basetypes.Float64Valuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type Float64ValuableWithValidateableAttribute interface {
	basetypes.Float64Valuable

	xattr.ValidateableAttribute
}

// Float64ValuableWithValidateableParameter extends the basetypes.Float64Valuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type Float64ValuableWithValidateableParameter interface {
	basetypes.Float64Valuable

	ValidateableParameter
}
