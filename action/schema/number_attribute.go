// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisifies the desired interfaces.
var (
	_ Attribute                               = NumberAttribute{}
	_ fwxschema.AttributeWithNumberValidators = NumberAttribute{}
)

// NumberAttribute represents a schema attribute that is a generic number with
// up to 512 bits of floating point or integer precision. When retrieving the
// value for this attribute, use types.Number as the value type unless the
// CustomType field is set.
//
// Use Float64Attribute for 64-bit floating point number attributes or
// Int64Attribute for 64-bit integer number attributes.
//
// Terraform configurations configure this attribute using expressions that
// return a number or directly via a floating point or integer value.
//
//	example_attribute = 123
//
// Terraform configurations reference this attribute using the attribute name.
//
//	.example_attribute
type NumberAttribute struct {
	// CustomType enables the use of a custom attribute type in place of the
	// default basetypes.NumberType. When retrieving data, the basetypes.NumberValuable
	// associated with this custom type must be used in place of types.Number.
	CustomType basetypes.NumberTypable

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose to enter a value
	// for this attribute or not. Optional and Required cannot both be true.
	Optional bool

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// DeprecationMessage defines warning diagnostic details to display when
	// practitioner configurations use this Attribute. The warning diagnostic
	// summary is automatically set to "Attribute Deprecated" along with
	// configuration source file and line information.
	//
	// Set this field to a practitioner actionable message such as:
	//
	//  - "Configure other_attribute instead. This attribute will be removed
	//    in the next major version of the provider."
	//  - "Remove this attribute's configuration as it no longer is used and
	//    the attribute will be removed in the next major version of the
	//    provider."
	//
	// In Terraform 1.2.7 and later, this warning diagnostic is displayed any
	// time a practitioner attempts to configure a value for this attribute and
	// certain scenarios where this attribute is referenced.
	//
	// In Terraform 1.2.6 and earlier, this warning diagnostic is only
	// displayed when the Attribute is Required or Optional, and if the
	// practitioner configuration sets the value to a known or unknown value
	// (which may eventually be null).
	//
	// Across any Terraform version, there are no warnings raised for
	// practitioner configuration values set directly to null, as there is no
	// way for the framework to differentiate between an unset and null
	// configuration due to how Terraform sends configuration information
	// across the protocol.
	//
	// Additional information about deprecation enhancements for read-only
	// attributes can be found in:
	//
	//  - https://github.com/hashicorp/terraform/issues/7569
	//
	DeprecationMessage string

	// Validators define value validation functionality for the attribute. All
	// elements of the slice of AttributeValidator are run, regardless of any
	// previous error diagnostics.
	//
	// Many common use case validators can be found in the
	// github.com/hashicorp/terraform-plugin-framework-validators Go module.
	//
	// If the Type field points to a custom type that implements the
	// xattr.TypeWithValidate interface, the validators defined in this field
	// are run in addition to the validation defined by the type.
	Validators []validator.Number
}

// ApplyTerraform5AttributePathStep always returns an error as it is not
// possible to step further into a NumberAttribute.
func (a NumberAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal returns true if the given Attribute is a NumberAttribute
// and all fields are equal.
func (a NumberAttribute) Equal(o fwschema.Attribute) bool {
	if _, ok := o.(NumberAttribute); !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (a NumberAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription returns the Description field value.
func (a NumberAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a NumberAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType returns types.NumberType or the CustomType field value if defined.
func (a NumberAttribute) GetType() attr.Type {
	if a.CustomType != nil {
		return a.CustomType
	}

	return types.NumberType
}

// IsComputed always returns false as action schema attributes cannot be Computed.
func (a NumberAttribute) IsComputed() bool {
	return false
}

// IsOptional returns the Optional field value.
func (a NumberAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a NumberAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive always returns false as action schema attributes cannot be Sensitive.
func (a NumberAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly always returns false as action schema attributes cannot be WriteOnly.
func (a NumberAttribute) IsWriteOnly() bool {
	return false
}

// IsRequiredForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a NumberAttribute) IsRequiredForImport() bool {
	return false
}

// IsOptionalForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a NumberAttribute) IsOptionalForImport() bool {
	return false
}

// NumberValidators returns the Validators field value.
func (a NumberAttribute) NumberValidators() []validator.Number {
	return a.Validators
}
