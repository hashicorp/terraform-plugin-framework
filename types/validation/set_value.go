// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SetValuableWithValidateableAttribute extends the basetypes.SetValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type SetValuableWithValidateableAttribute interface {
	basetypes.SetValuable

	xattr.ValidateableAttribute
}

// SetValuableWithValidateableParameter extends the basetypes.SetValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type SetValuableWithValidateableParameter interface {
	basetypes.SetValuable

	ValidateableParameter
}
