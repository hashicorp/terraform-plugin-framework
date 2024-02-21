// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// DefaultParameterName is the name given to parameters which do not declare
// a name. Use this to prevent Terraform errors for missing names.
const DefaultParameterName = "param"

// Parameter is the interface for defining function parameters.
type Parameter interface {
	// GetAllowNullValue should return if the parameter accepts a null value.
	GetAllowNullValue() bool

	// GetAllowUnknownValues should return if the parameter accepts an unknown
	// value.
	GetAllowUnknownValues() bool

	// GetDescription should return the plaintext documentation for the
	// parameter.
	GetDescription() string

	// GetMarkdownDescription should return the Markdown documentation for the
	// parameter.
	GetMarkdownDescription() string

	// GetName should return a usage name for the parameter. Parameters are
	// positional, so this name has no meaning except documentation.
	GetName() string

	// GetType should return the data type for the parameter, which determines
	// what data type Terraform requires for configurations setting the argument
	// during a function call and the argument data type received by the
	// Function type Run method.
	GetType() attr.Type
}
