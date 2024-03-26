// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// NumberValuableWithValidateableAttribute extends the basetypes.NumberValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type NumberValuableWithValidateableAttribute interface {
	basetypes.NumberValuable

	xattr.ValidateableAttribute
}

// NumberValuableWithValidateableParameter extends the basetypes.NumberValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type NumberValuableWithValidateableParameter interface {
	basetypes.NumberValuable

	ValidateableParameter
}
