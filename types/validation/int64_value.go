// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Int64ValuableWithValidateableAttribute extends the basetypes.Int64Valuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type Int64ValuableWithValidateableAttribute interface {
	basetypes.Int64Valuable

	xattr.ValidateableAttribute
}

// Int64ValuableWithValidateableParameter extends the basetypes.Int64Valuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type Int64ValuableWithValidateableParameter interface {
	basetypes.Int64Valuable

	ValidateableParameter
}
