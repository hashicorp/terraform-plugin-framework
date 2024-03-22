// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ListValuableWithValidateableAttribute extends the basetypes.ListValuable interface to include a
// xattr.ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type ListValuableWithValidateableAttribute interface {
	basetypes.ListValuable

	xattr.ValidateableAttribute
}

// ListValuableWithValidateableParameter extends the basetypes.ListValuable interface to include a
// ValidateableParameter interface, used to bundle consistent parameter validation logic with
// the Value.
type ListValuableWithValidateableParameter interface {
	basetypes.ListValuable

	ValidateableParameter
}
