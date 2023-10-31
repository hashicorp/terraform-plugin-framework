// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ Parameter = ObjectParameter{}

// ObjectParameter represents a function parameter that is a mapping of
// defined attribute names to values. Either the AttributeTypes or CustomType
// field must be set.
//
// When retrieving the argument value for this parameter:
//
//   - If CustomType is set, use its associated value type.
//   - If AllowUnknownValues is enabled, you must use the [types.Object] value
//     type.
//   - If AllowNullValue is enabled, you must use the [types.Object] or a
//     compatible Go *struct value type.
//   - Otherwise, use [types.Object] or compatible *struct/struct value types.
//
// Terraform configurations set this parameter's argument data using expressions
// that return an object or directly via object ("{...}") syntax.
type ObjectParameter struct {
	// AttributeTypes is the mapping of underlying attribute names to attribute
	// types. This field must be set.
	AttributeTypes map[string]attr.Type

	// AllowNullValue when enabled denotes that a null argument value can be
	// passed to the function. When disabled, Terraform returns an error if the
	// argument value is null.
	AllowNullValue bool

	// AllowUnknownValues when enabled denotes that an unknown argument value
	// can be passed to the function. When disabled, Terraform skips the
	// function call entirely and assumes an unknown value result from the
	// function.
	AllowUnknownValues bool

	// CustomType enables the use of a custom data type in place of the
	// default [basetypes.ObjectType]. When retrieving data, the
	// [basetypes.ObjectValuable] implementation associated with this custom
	// type must be used in place of [types.Object].
	CustomType basetypes.ObjectTypable

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this parameter is,
	// what it is for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this parameter is, what it is for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// Name is a short usage name for the parameter, such as "data". This name
	// is used in documentation, such as generating a function signature,
	// however its usage may be extended in the future. If no name is provided,
	// this will default to "param".
	//
	// This must be a valid Terraform identifier, such as starting with an
	// alphabetical character and followed by alphanumeric or underscore
	// characters.
	Name string
}

// GetAllowNullValue returns if the parameter accepts a null value.
func (p ObjectParameter) GetAllowNullValue() bool {
	return p.AllowNullValue
}

// GetAllowUnknownValues returns if the parameter accepts an unknown value.
func (p ObjectParameter) GetAllowUnknownValues() bool {
	return p.AllowUnknownValues
}

// GetDescription returns the parameter plaintext description.
func (p ObjectParameter) GetDescription() string {
	return p.Description
}

// GetMarkdownDescription returns the parameter Markdown description.
func (p ObjectParameter) GetMarkdownDescription() string {
	return p.MarkdownDescription
}

// GetName returns the parameter name.
func (p ObjectParameter) GetName() string {
	if p.Name != "" {
		return p.Name
	}

	return DefaultParameterName
}

// GetType returns the parameter data type.
func (p ObjectParameter) GetType() attr.Type {
	if p.CustomType != nil {
		return p.CustomType
	}

	return basetypes.ObjectType{
		AttrTypes: p.AttributeTypes,
	}
}
