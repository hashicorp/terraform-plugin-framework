// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Float64AttributeWithValidate extends the basetypes.Float64Valuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type Float64AttributeWithValidate interface {
	basetypes.Float64Valuable

	ValidateableAttribute
}
