// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// StringValuableWithValidateableAttribute extends the basetypes.StringValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type StringValuableWithValidateableAttribute interface {
	basetypes.StringValuable

	xattr.ValidateableAttribute
}

// StringValuableWithValidateableParameter extends the basetypes.StringValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type StringValuableWithValidateableParameter interface {
	basetypes.StringValuable

	ValidateableParameter
}
