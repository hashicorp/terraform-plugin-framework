// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// StringAttributeWithValidate extends the basetypes.StringValuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type StringAttributeWithValidate interface {
	basetypes.StringValuable

	ValidateableAttribute
}
