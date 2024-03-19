// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// BoolAttributeWithValidate extends the basetypes.BoolValuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type BoolAttributeWithValidate interface {
	basetypes.BoolValuable

	ValidateableAttribute
}
