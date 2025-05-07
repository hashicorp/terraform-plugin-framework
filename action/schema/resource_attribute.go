// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisifies the desired interfaces.
var (
	_ Attribute                                    = ResourceAttribute{}
	_ fwschema.AttributeWithValidateImplementation = ResourceAttribute{}
)

// ResourceAttribute represents a schema attribute that is an Resource with only
// type information for underlying attributes. When retrieving the value for
// this attribute, use types.Resource as the value type unless the CustomType
// field is set. The AttributeTypes field must be set.
//
// Prefer SingleNestedAttribute over ResourceAttribute if the provider is
// using protocol version 6 and full attribute functionality is needed.
//
// Terraform configurations configure this attribute using expressions that
// return an Resource or directly via curly brace syntax.
//
//	# Resource with one attribute
//	example_attribute = {
//		underlying_attribute = #...
//	}
//
// Terraform configurations reference this attribute using expressions that
// accept an Resource or an attribute directly via period syntax:
//
//	# underlying attribute
//	.example_attribute.underlying_attribute
type ResourceAttribute struct {
	Resource resource.Resource

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
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
}

// ApplyTerraform5AttributePathStep returns the result of stepping into an
// attribute name or an error.
func (a ResourceAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal returns true if the given Attribute is a ResourceAttribute
// and all fields are equal.
func (a ResourceAttribute) Equal(o fwschema.Attribute) bool {
	if _, ok := o.(ResourceAttribute); !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (a ResourceAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription returns the Description field value.
func (a ResourceAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a ResourceAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType returns types.ResourceType or the CustomType field value if defined.
func (a ResourceAttribute) GetType() attr.Type {
	schemaReq := resource.SchemaRequest{}
	schemaResp := resource.SchemaResponse{}

	a.Resource.Schema(context.TODO(), schemaReq, &schemaResp)
	return schemaResp.Schema.Type()
}

// IsComputed returns the Computed field value.
func (a ResourceAttribute) IsComputed() bool {
	return false
}

// IsOptional returns the Optional field value.
func (a ResourceAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a ResourceAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive returns the Sensitive field value.
func (a ResourceAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly returns false as write-only attributes are not relevant to ephemeral resource schemas,
// as these schemas describe data that is explicitly not saved to any artifact.
func (a ResourceAttribute) IsWriteOnly() bool {
	return false
}

// IsRequiredForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a ResourceAttribute) IsRequiredForImport() bool {
	return false
}

// IsOptionalForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a ResourceAttribute) IsOptionalForImport() bool {
	return false
}

// ValidateImplementation contains logic for validating the
// provider-defined implementation of the attribute to prevent unexpected
// errors or panics. This logic runs during the GetProviderSchema RPC
// and should never include false positives.
func (a ResourceAttribute) ValidateImplementation(ctx context.Context, req fwschema.ValidateImplementationRequest, resp *fwschema.ValidateImplementationResponse) {
	if a.Resource == nil {
		// TODO: create proper error
		resp.Diagnostics.Append(fwschema.AttributeMissingAttributeTypesDiag(req.Path))
	}

	if fwtype.ContainsCollectionWithDynamic(a.GetType()) {
		resp.Diagnostics.Append(fwtype.AttributeCollectionWithDynamicTypeDiag(req.Path))
	}
}
