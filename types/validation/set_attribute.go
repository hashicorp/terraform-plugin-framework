// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SetAttributeWithValidate extends the basetypes.SetValuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type SetAttributeWithValidate interface {
	basetypes.SetValuable

	xattr.ValidateableAttribute
}
