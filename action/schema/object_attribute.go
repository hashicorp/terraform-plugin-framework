// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisifies the desired interfaces.
var (
	_ Attribute                                    = ObjectAttribute{}
	_ fwschema.AttributeWithValidateImplementation = ObjectAttribute{}
	_ fwxschema.AttributeWithObjectValidators      = ObjectAttribute{}
)

// ObjectAttribute represents a schema attribute that is an object with only
// type information for underlying attributes. When retrieving the value for
// this attribute, use types.Object as the value type unless the CustomType
// field is set. The AttributeTypes field must be set.
//
// Prefer SingleNestedAttribute over ObjectAttribute if the provider is
// using protocol version 6 and full attribute functionality is needed.
//
// Terraform configurations configure this attribute using expressions that
// return an object or directly via curly brace syntax.
//
//	# object with one attribute
//	example_attribute = {
//		underlying_attribute = #...
//	}
//
// Terraform configurations reference this attribute using expressions that
// accept an object or an attribute directly via period syntax:
//
//	# underlying attribute
//	.example_attribute.underlying_attribute
type ObjectAttribute struct {
	// AttributeTypes is the mapping of underlying attribute names to attribute
	// types. This field must be set.
	//
	// Attribute types that contain a collection with a nested dynamic type (i.e. types.List[types.Dynamic]) are not supported.
	// If underlying dynamic collection values are required, replace this attribute definition with
	// DynamicAttribute instead.
	AttributeTypes map[string]attr.Type

	// CustomType enables the use of a custom attribute type in place of the
	// default basetypes.ObjectType. When retrieving data, the basetypes.ObjectValuable
	// associated with this custom type must be used in place of types.Object.
	CustomType basetypes.ObjectTypable

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
	Validators []validator.Object
}

// ApplyTerraform5AttributePathStep returns the result of stepping into an
// attribute name or an error.
func (a ObjectAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal returns true if the given Attribute is a ObjectAttribute
// and all fields are equal.
func (a ObjectAttribute) Equal(o fwschema.Attribute) bool {
	if _, ok := o.(ObjectAttribute); !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (a ObjectAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription returns the Description field value.
func (a ObjectAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a ObjectAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType returns types.ObjectType or the CustomType field value if defined.
func (a ObjectAttribute) GetType() attr.Type {
	if a.CustomType != nil {
		return a.CustomType
	}

	return types.ObjectType{
		AttrTypes: a.AttributeTypes,
	}
}

// IsComputed always returns false as action schema attributes cannot be Computed.
func (a ObjectAttribute) IsComputed() bool {
	return false
}

// IsOptional returns the Optional field value.
func (a ObjectAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a ObjectAttribute) IsRequired() bool {
	return a.Required
}

// IsWriteOnly always returns false as action schema attributes cannot be WriteOnly.
func (a ObjectAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly always returns false as action schema attributes cannot be WriteOnly.
func (a ObjectAttribute) IsWriteOnly() bool {
	return false
}

// IsRequiredForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a ObjectAttribute) IsRequiredForImport() bool {
	return false
}

// IsOptionalForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a ObjectAttribute) IsOptionalForImport() bool {
	return false
}

// ObjectValidators returns the Validators field value.
func (a ObjectAttribute) ObjectValidators() []validator.Object {
	return a.Validators
}

// ValidateImplementation contains logic for validating the
// provider-defined implementation of the attribute to prevent unexpected
// errors or panics. This logic runs during the GetProviderSchema RPC
// and should never include false positives.
func (a ObjectAttribute) ValidateImplementation(ctx context.Context, req fwschema.ValidateImplementationRequest, resp *fwschema.ValidateImplementationResponse) {
	if a.AttributeTypes == nil && a.CustomType == nil {
		resp.Diagnostics.Append(fwschema.AttributeMissingAttributeTypesDiag(req.Path))
	}

	if a.CustomType == nil && fwtype.ContainsCollectionWithDynamic(a.GetType()) {
		resp.Diagnostics.Append(fwtype.AttributeCollectionWithDynamicTypeDiag(req.Path))
	}
}
