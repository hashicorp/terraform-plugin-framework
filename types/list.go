package types

import (
	"context"

	tf "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ListType struct {
	ElemType tf.AttributeType
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (l ListType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.List{
		ElementType: l.ElemType.TerraformType(ctx),
	}
}

// Validate returns any warnings or errors about the value that is
// being used to populate the AttributeType. It is generally used to
// check the data format and ensure that it complies with the
// requirements of the AttributeType.
//
// TODO: don't use tfprotov6.Diagnostic, use our type
func (l ListType) Validate(_ context.Context, _ tftypes.Value) []*tfprotov6.Diagnostic {
	return nil
}

// Description returns a practitioner-friendly explanation of the type
// and the constraints of the data it accepts and returns. It will be
// combined with the Description associated with the Attribute.
func (l ListType) Description(_ context.Context, _ tf.StringKind) string {
	return ""
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (l ListType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (tf.AttributeValue, error) {
	if !in.IsKnown() {
		return List{
			Unknown: true,
		}, nil
	}
	if in.IsNull() {
		return List{
			Null: true,
		}, nil
	}
	val := []tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	elems := make([]tf.AttributeValue, 0, len(val))
	for _, elem := range val {
		av, err := l.ElemType.ValueFromTerraform(ctx, elem)
		if err != nil {
			return nil, err
		}
		elems = append(elems, av)
	}
	return List{
		Elems:    elems,
		ElemType: l.TerraformType(ctx),
	}, nil
}

type List struct {
	Unknown  bool
	Null     bool
	Elems    []tf.AttributeValue
	ElemType tftypes.Type
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (l List) ToTerraformValue(ctx context.Context) (interface{}, error) {
	if l.Unknown {
		return tftypes.UnknownValue, nil
	}
	if l.Null {
		return nil, nil
	}
	vals := make([]tftypes.Value, 0, len(l.Elems))
	for _, elem := range l.Elems {
		val, err := elem.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		err = tftypes.ValidateValue(l.ElemType, val)
		if err != nil {
			return nil, err
		}
		vals = append(vals, tftypes.NewValue(l.ElemType, val))
	}
	return vals, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (l List) Equal(o tf.AttributeValue) bool {
	other, ok := o.(List)
	if !ok {
		return false
	}
	if l.Unknown != other.Unknown {
		return false
	}
	if l.Null != other.Null {
		return false
	}
	if !l.ElemType.Is(other.ElemType) {
		return false
	}
	if len(l.Elems) != len(other.Elems) {
		return false
	}
	for pos, lElem := range l.Elems {
		oElem := other.Elems[pos]
		if !lElem.Equal(oElem) {
			return false
		}
	}
	return true
}
