// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xattr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// TypeWithValidate extends the attr.Type interface to include a Validate
// method, used to bundle consistent validation logic with the Type.
//
// Deprecated: Use the value type-specific AttributeWithValidate and
// ParameterWithValidate interfaces in the validation package instead.
type TypeWithValidate interface {
	attr.Type

	// Validate returns any warnings or errors about the value that is
	// being used to populate the Type. It is generally used to check the
	// data format and ensure that it complies with the requirements of the
	// Type.
	Validate(context.Context, tftypes.Value, path.Path) diag.Diagnostics
}
