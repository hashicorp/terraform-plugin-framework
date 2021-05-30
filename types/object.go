package types

import (
	"context"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ObjectType is an AttributeType representing a object
type ObjectType struct {
	AttributeTypes map[string]tfsdk.AttributeType
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (o ObjectType) TerraformType(ctx context.Context) tftypes.Type {
	var attributeTypes map[string]tftypes.Type
	for k, v := range o.AttributeTypes {
		attributeTypes[k] = v.TerraformType(ctx)
	}
	return tftypes.Object{
		AttributeTypes: attributeTypes,
	}
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (o ObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	object := &Object{
		AttributeTypes: o.AttributeTypes,
	}
	err := object.SetTerraformValue(ctx, in)
	return object, err
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

	Attributes map[string]tfsdk.AttributeValue

	AttributeTypes map[string]tfsdk.AttributeType
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (o *Object) ToTerraformValue(ctx context.Context) (interface{}, error) {
	if o.Unknown {
		return tftypes.UnknownValue, nil
	}
	if o.Null {
		return nil, nil
	}
	var vals map[string]tftypes.Value

	for k, v := range o.Attributes {
		val, err := v.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		err = tftypes.ValidateValue(o.AttributeTypes[k].TerraformType(ctx), val)
		if err != nil {
			return nil, err
		}
		vals[k] = tftypes.NewValue(o.AttributeTypes[k].TerraformType(ctx), val)
	}
	return vals, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (o *Object) Equal(c tfsdk.AttributeValue) bool {
	other, ok := c.(*Object)
	if !ok {
		return false
	}
	if o.Unknown != other.Unknown {
		return false
	}
	if o.Null != other.Null {
		return false
	}
	if len(o.AttributeTypes) != len(other.AttributeTypes) {
		return false
	}
	for k, v := range o.AttributeTypes {
		attr, ok := other.AttributeTypes[k]
		if !ok {
			return false
		}
		if !v.Equal(attr) {
			return false
		}
	}
	if len(o.Attributes) != len(other.Attributes) {
		return false
	}
	for k, v := range o.Attributes {
		attr, ok := other.Attributes[k]
		if !ok {
			return false
		}
		if !v.Equal(attr) {
			return false
		}
	}

	return true
}

func (o *Object) SetTerraformValue(ctx context.Context, in tftypes.Value) error {
	o.Unknown = false
	o.Null = false
	o.Attributes = nil
	if !in.IsKnown() {
		o.Unknown = true
		return nil
	}
	if in.IsNull() {
		o.Null = true
		return nil
	}
	attributes := map[string]tfsdk.AttributeValue{}

	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return err
	}

	for k, v := range val {
		a, err := o.AttributeTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return err
		}
		attributes[k] = a
	}
	o.Attributes = attributes
	return nil
}
