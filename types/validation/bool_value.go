// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// BoolValuableWithValidateableAttribute extends the basetypes.BoolValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type BoolValuableWithValidateableAttribute interface {
	basetypes.BoolValuable

	xattr.ValidateableAttribute
}

// BoolValuableWithValidateableParameter extends the basetypes.BoolValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type BoolValuableWithValidateableParameter interface {
	basetypes.BoolValuable

	ValidateableParameter
}
