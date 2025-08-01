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
	_ Attribute                                = DynamicAttribute{}
	_ fwxschema.AttributeWithDynamicValidators = DynamicAttribute{}
)

// DynamicAttribute represents a schema attribute that is a dynamic, rather
// than a single static type. Static types are always preferable over dynamic
// types in Terraform as practitioners will receive less helpful configuration
// assistance from validation error diagnostics and editor integrations. When
// retrieving the value for this attribute, use types.Dynamic as the value type
// unless the CustomType field is set.
//
// The concrete value type for a dynamic is determined at runtime by Terraform, if defined in the configuration.
type DynamicAttribute struct {
	// CustomType enables the use of a custom attribute type in place of the
	// default basetypes.DynamicType. When retrieving data, the basetypes.DynamicValuable
	// associated with this custom type must be used in place of types.Dynamic.
	CustomType basetypes.DynamicTypable

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
	// (which may eventually be null). It has no effect when the Attribute is
	// Computed-only (read-only; not Required or Optional).
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
	Validators []validator.Dynamic
}

// ApplyTerraform5AttributePathStep always returns an error as it is not
// possible to step further into a DynamicAttribute.
func (a DynamicAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal returns true if the given Attribute is a DynamicAttribute
// and all fields are equal.
func (a DynamicAttribute) Equal(o fwschema.Attribute) bool {
	if _, ok := o.(DynamicAttribute); !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (a DynamicAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription returns the Description field value.
func (a DynamicAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a DynamicAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType returns types.DynamicType or the CustomType field value if defined.
func (a DynamicAttribute) GetType() attr.Type {
	if a.CustomType != nil {
		return a.CustomType
	}

	return types.DynamicType
}

// IsComputed returns false because it does not apply to ListResource schemas.
func (a DynamicAttribute) IsComputed() bool {
	return false
}

// IsOptional returns the Optional field value.
func (a DynamicAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a DynamicAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive returns false because it does not apply to ListResource schemas.
func (a DynamicAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly returns false as write-only attributes are not relevant to ephemeral resource schemas,
// as these schemas describe data that is explicitly not saved to any artifact.
func (a DynamicAttribute) IsWriteOnly() bool {
	return false
}

// DynamicValidators returns the Validators field value.
func (a DynamicAttribute) DynamicValidators() []validator.Dynamic {
	return a.Validators
}

// IsRequiredForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a DynamicAttribute) IsRequiredForImport() bool {
	return false
}

// IsOptionalForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a DynamicAttribute) IsOptionalForImport() bool {
	return false
}
