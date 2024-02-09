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
	attr.Type

	// ValueFromDynamic should convert the DynamicValue to a DynamicValuable type.
	ValueFromDynamic(context.Context, DynamicValue) (DynamicValuable, diag.Diagnostics)
}

var _ DynamicTypable = DynamicType{}

// TODO: doc
type DynamicType struct{}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the type.
func (t DynamicType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	// TODO: Need to verify this won't cause issues elsewhere
	//
	// Based on dynamic type alone, there is no alternative type information to return related to a path step.
	// However, it is possible for a dynamic type to have an underlying value of a list, set, map, object, or tuple
	// that will have corresponding path steps. the type
	// will have no reference to this value.
	//
	// Since the dynamic type has no reference to the underlying value, we just return the dynamic type which can be used
	// to grab the attr.Value from `(DynamicType).ValueFromTerraform`; which can be used to create any underlying value.
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

	// TODO: is it possible to receive a DynamicPseudoType with a known value from terraform-plugin-go or Terraform core? I don't think so
	// - tftypes.NewValue() does allow you to create known values with DynamicPseudoType, but I'm not sure if that is really possible in normal
	// 	 Terraform operations.
	// - (tfprotov6.DynamicValue).Unmarshal only uses DynamicPseudoType for "dynamic" marked values from Terraform. I believe this only happens
	// 	 with null and unknown values where the schema is DynamicPseudoType.
	// If it is possible, there will be an infinite recursive bug without this if statement
	if in.Type().Is(tftypes.DynamicPseudoType) {
		return nil, errors.New("ambiguous known value for `tftypes.DynamicPseudoType` detected")
	}

	attrType := tfToAttr(in.Type())
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
func tfToAttr(in tftypes.Type) attr.Type {
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

		elemType := tfToAttr(l.ElementType)
		return ListType{ElemType: elemType}
	}
	if in.Is(tftypes.Map{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		m := in.(tftypes.Map)

		elemType := tfToAttr(m.ElementType)
		return MapType{ElemType: elemType}
	}
	if in.Is(tftypes.Set{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		s := in.(tftypes.Set)

		elemType := tfToAttr(s.ElementType)
		return SetType{ElemType: elemType}
	}

	// Structural types
	if in.Is(tftypes.Object{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		o := in.(tftypes.Object)

		attrTypes := make(map[string]attr.Type, len(o.AttributeTypes))
		for name, tfType := range o.AttributeTypes {
			attrTypes[name] = tfToAttr(tfType)
		}
		return ObjectType{AttrTypes: attrTypes}
	}
	if in.Is(tftypes.Tuple{}) {
		//nolint:forcetypeassert // Type assertion is guaranteed by the above `(tftypes.Type).Is` function
		t := in.(tftypes.Tuple)

		elemTypes := make([]attr.Type, len(t.ElementTypes))
		for i, tfType := range t.ElementTypes {
			elemTypes[i] = tfToAttr(tfType)
		}
		return TupleType{ElemTypes: elemTypes}
	}

	// TODO: probably return an error to bubble up
	panic("need to handle this")
}
