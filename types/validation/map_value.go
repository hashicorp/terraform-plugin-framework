// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MapValuableWithValidateableAttribute extends the basetypes.MapValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type MapValuableWithValidateableAttribute interface {
	basetypes.MapValuable

	xattr.ValidateableAttribute
}

// MapValuableWithValidateableParameter extends the basetypes.MapValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type MapValuableWithValidateableParameter interface {
	basetypes.MapValuable

	ValidateableParameter
}
