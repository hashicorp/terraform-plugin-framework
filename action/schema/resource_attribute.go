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
	// Resource is the framework resource that constitutes
	// the type of this attribute. The resource's Schema() func
	// will be called to create type for the attribute.
	Resource resource.Resource

	Required bool

	tftypes.Type

	Optional bool

	Description string

	MarkdownDescription string

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
