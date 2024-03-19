// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Int64AttributeWithValidate extends the basetypes.Int64Valuable interface to include a
// ValidateableAttribute interface, used to bundle consistent attribute validation logic with
// the Value.
type Int64AttributeWithValidate interface {
	basetypes.Int64Valuable

	ValidateableAttribute
}
