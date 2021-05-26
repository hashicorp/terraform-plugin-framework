package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ObjectType is an AttributeType representing a object
type ObjectType struct {
	AttributeTypes map[string]attr.Type
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (l ObjectType) TerraformType(ctx context.Context) tftypes.Type {
	var attributeTypes map[string]tftypes.Type
	for k, v := range l.AttributeTypes {
		attributeTypes[k] = v.TerraformType(ctx)
	}
	return tftypes.Object{
		AttributeTypes: attributeTypes,
	}
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (l ObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Object{
			Unknown: true,
		}, nil
	}
	if in.IsNull() {
		return Object{
			Null: true,
		}, nil
	}
	var attributes map[string]attr.Value

	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := l.AttributeTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}

	var attributeTypes map[string]tftypes.Type
	for k, v := range l.AttributeTypes {
		attributeTypes[k] = v.TerraformType(ctx)
	}

	return Object{
		Attributes:     attributes,
		AttributeTypes: attributeTypes,
	}, nil
}

// Object represents an object
type Object struct {
	// Unknown will be set to true if the entire object is an unknown value.
	// If only some of the elements in the object are unknown, their known or
	// unknown status will be represented however that AttributeValue
	// surfaces that information. The Object's Unknown property only tracks
	// if the number of elements in a Object is known, not whether the
	// elements that are in the object are known.
	Unknown bool

	// Null will be set to true if the object is null, either because it was
	// omitted from the configuration, state, or plan, or because it was
	// explicitly set to null.
	Null bool

	Attributes map[string]attr.Value

	AttributeTypes map[string]tftypes.Type
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (l Object) ToTerraformValue(ctx context.Context) (interface{}, error) {
	if l.Unknown {
		return tftypes.UnknownValue, nil
	}
	if l.Null {
		return nil, nil
	}
	var vals map[string]tftypes.Value

	for k, v := range l.Attributes {
		val, err := v.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		err = tftypes.ValidateValue(l.AttributeTypes[k], val)
		if err != nil {
			return nil, err
		}
		vals[k] = tftypes.NewValue(l.AttributeTypes[k], val)
	}
	return vals, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (l Object) Equal(o attr.Value) bool {
	other, ok := o.(Object)
	if !ok {
		return false
	}
	if l.Unknown != other.Unknown {
		return false
	}
	if l.Null != other.Null {
		return false
	}
	// TODO this properly
	for k, v := range l.AttributeTypes {
		if !v.Is(other.AttributeTypes[k]) {
			return false
		}
	}
	for k, v := range other.AttributeTypes {
		if !v.Is(l.AttributeTypes[k]) {
			return false
		}
	}
	for k, v := range l.Attributes {
		if !v.Equal(other.Attributes[k]) {
			return false
		}
	}
	for k, v := range other.Attributes {
		if !v.Equal(l.Attributes[k]) {
			return false
		}
	}

	return true
}
