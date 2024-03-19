// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// MapAttributeWithValidate extends the basetypes.MapValuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type MapAttributeWithValidate interface {
	basetypes.MapValuable

	ValidateableAttribute
}
