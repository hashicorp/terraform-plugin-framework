// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// DynamicTypable extends attr.Type for dynamic types. Implement this interface to create a custom DynamicType type.
type DynamicTypable interface {
	attr.TypeWithDynamicValue

	// ValueFromDynamic should convert the DynamicValue to a DynamicValuable type.
	ValueFromDynamic(context.Context, DynamicValue) (DynamicValuable, diag.Diagnostics)
}

var _ DynamicTypable = DynamicType{}

// TODO: doc
type DynamicType struct{}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the type.
func (t DynamicType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	// MAINTAINER NOTE: Based on dynamic type alone, there is no alternative type information to return related to a path step.
	// However, it is possible for a dynamic type to have an underlying value of a list, set, map, object, or tuple
	// that will have corresponding path steps.
	//
	// Since the dynamic type has no reference to the underlying value, we just return the dynamic type which can be used
	// to grab the attr.Value from `(DynamicType).ValueFromTerraform`.
	return t, nil
}

// Equal returns true if the given type is equivalent.
func (t DynamicType) Equal(o attr.Type) bool {
	_, ok := o.(DynamicType)

	return ok
}

// String returns a human-friendly description of the DynamicType.
func (t DynamicType) String() string {
	return "basetypes.DynamicType"
}

// TerraformType returns the tftypes.Type that should be used to represent this type.
func (t DynamicType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.DynamicPseudoType
}

// ValueFromDynamic returns a DynamicValuable type given a DynamicValue.
func (t DynamicType) ValueFromDynamic(ctx context.Context, v DynamicValue) (DynamicValuable, diag.Diagnostics) {
	return v, nil
}

// ValueFromTerraform returns an attr.Value given a tftypes.Value. This is meant to convert
// the tftypes.Value into a more convenient Go type for the provider to consume the data with.
func (t DynamicType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return NewDynamicUnknown(), nil
	}

	if in.IsNull() {
		return NewDynamicNull(), nil
	}

	// MAINTAINER NOTE: It should not be possible for Terraform core to send a known value of `tftypes.DynamicPseudoType`.
	// This check prevents an infinite recursion when attempting to create a dynamic value.
	if in.Type().Is(tftypes.DynamicPseudoType) {
		return nil, errors.New("ambiguous known value for `tftypes.DynamicPseudoType` detected")
	}

	attrType := t.DetermineAttrType(in.Type())
	val, err := attrType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	return NewDynamicValue(val), nil
}

// ValueType returns the Value type.
func (t DynamicType) ValueType(_ context.Context) attr.Value {
	return DynamicValue{}
}

// TODO: move this somewhere else?
func (t DynamicType) DetermineAttrType(in tftypes.Type) attr.Type {
	// Primitive types
	if in.Is(tftypes.Bool) {
		return BoolType{}
	}
	if in.Is(tftypes.Number) {
		return NumberType{}
	}
	if in.Is(tftypes.String) {
		return StringType{}
	}
	if in.Is(tftypes.DynamicPseudoType) {
		// Null and Unknown values that do not have a type determined will have a type of DynamicPseudoType
		return DynamicType{}
	}

	// Collection types
	if in.Is(tftypes.List{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		l := in.(tftypes.List)

		elemType := t.DetermineAttrType(l.ElementType)
		return ListType{ElemType: elemType}
	}
	if in.Is(tftypes.Map{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		m := in.(tftypes.Map)

		elemType := t.DetermineAttrType(m.ElementType)
		return MapType{ElemType: elemType}
	}
	if in.Is(tftypes.Set{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		s := in.(tftypes.Set)

		elemType := t.DetermineAttrType(s.ElementType)
		return SetType{ElemType: elemType}
	}

	// Structural types
	if in.Is(tftypes.Object{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		o := in.(tftypes.Object)

		attrTypes := make(map[string]attr.Type, len(o.AttributeTypes))
		for name, tfType := range o.AttributeTypes {
			attrTypes[name] = t.DetermineAttrType(tfType)
		}
		return ObjectType{AttrTypes: attrTypes}
	}
	if in.Is(tftypes.Tuple{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		tup := in.(tftypes.Tuple)

		elemTypes := make([]attr.Type, len(tup.ElementTypes))
		for i, tfType := range tup.ElementTypes {
			elemTypes[i] = t.DetermineAttrType(tfType)
		}
		return TupleType{ElemTypes: elemTypes}
	}

	// TODO: I think it'd probably be best to return an error from this?
	panic("need to handle this")
}