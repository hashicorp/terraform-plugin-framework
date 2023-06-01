// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ fwschema.NestedAttribute = NestedAttribute{}

type NestedAttribute struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	NestingMode         fwschema.NestingMode
	Optional            bool
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	switch a.GetNestingMode() {
	case fwschema.NestingModeList:
		_, ok := step.(tftypes.ElementKeyInt)

		if !ok {
			return nil, fmt.Errorf("cannot apply step %T to ListNestedAttribute", step)
		}

		return a.NestedObject, nil
	case fwschema.NestingModeMap:
		_, ok := step.(tftypes.ElementKeyString)

		if !ok {
			return nil, fmt.Errorf("cannot apply step %T to ListNestedAttribute", step)
		}

		return a.NestedObject, nil
	case fwschema.NestingModeSet:
		_, ok := step.(tftypes.ElementKeyValue)

		if !ok {
			return nil, fmt.Errorf("cannot apply step %T to ListNestedAttribute", step)
		}

		return a.NestedObject, nil
	case fwschema.NestingModeSingle:
		name, ok := step.(tftypes.AttributeName)

		if !ok {
			return nil, fmt.Errorf("cannot apply step %T to SingleNestedAttribute", step)
		}

		attribute, ok := a.GetNestedObject().GetAttributes()[string(name)]

		if !ok {
			return nil, fmt.Errorf("no attribute %q on SingleNestedAttribute", name)
		}

		return attribute, nil
	default:
		panic(fmt.Sprintf("nesting mode not supported: %T", a.GetNestingMode()))
	}
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttribute) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttribute)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttribute) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttribute) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttribute) GetNestingMode() fwschema.NestingMode {
	return a.NestingMode
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttribute) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	switch a.GetNestingMode() {
	case fwschema.NestingModeList:
		return types.ListType{
			ElemType: a.GetNestedObject().Type(),
		}
	case fwschema.NestingModeMap:
		return types.MapType{
			ElemType: a.GetNestedObject().Type(),
		}
	case fwschema.NestingModeSet:
		return types.SetType{
			ElemType: a.GetNestedObject().Type(),
		}
	case fwschema.NestingModeSingle:
		return a.GetNestedObject().Type()
	default:
		return nil
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttribute) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttribute) IsSensitive() bool {
	return a.Sensitive
}
