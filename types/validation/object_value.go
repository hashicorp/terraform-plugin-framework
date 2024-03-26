// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ObjectValuableWithValidateableAttribute extends the basetypes.ObjectValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type ObjectValuableWithValidateableAttribute interface {
	basetypes.ObjectValuable

	xattr.ValidateableAttribute
}

// ObjectValuableWithValidateableParameter extends the basetypes.ObjectValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type ObjectValuableWithValidateableParameter interface {
	basetypes.ObjectValuable

	ValidateableParameter
}
