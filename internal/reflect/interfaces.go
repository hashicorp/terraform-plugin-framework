package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Unknownable is an interface for types that can be explicitly set to known or
// unknown.
type Unknownable interface {
	SetUnknown(context.Context, bool) error
	SetValue(context.Context, interface{}) error
	GetUnknown(context.Context) bool
	GetValue(context.Context) interface{}
}

// NewUnknownable creates a zero value of `target` (or the concrete type it's
// referencing, if it's a pointer) and calls its SetUnknown method.
//
// It is meant to be called through Into, not directly.
func NewUnknownable(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	receiver := pointerSafeZeroValue(ctx, target)
	method := receiver.MethodByName("SetUnknown")
	if !method.IsValid() {
		return target, path.NewErrorf("unexpectedly couldn't find SetUnknown method on type %s", receiver.Type().String())
	}
	results := method.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(!val.IsKnown()),
	})
	err := results[0].Interface()
	if err != nil {
		return target, path.NewError(err.(error))
	}
	return receiver, nil
}

// FromUnknownable creates an attr.Value from the data in an Unknownable.
//
// It is meant to be called through OutOf, not directly.
func FromUnknownable(ctx context.Context, typ attr.Type, val Unknownable, path *tftypes.AttributePath) (attr.Value, error) {
	if val.GetUnknown(ctx) {
		res, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), tftypes.UnknownValue))
		if err != nil {
			return nil, path.NewError(err)
		}
		return res, nil
	}
	err := tftypes.ValidateValue(typ.TerraformType(ctx), val.GetValue(ctx))
	if err != nil {
		return nil, path.NewError(err)
	}
	res, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), val.GetValue(ctx)))
	if err != nil {
		return nil, path.NewError(err)
	}
	return res, nil
}

// Nullable is an interface for types that can be explicitly set to null.
type Nullable interface {
	SetNull(context.Context, bool) error
	SetValue(context.Context, interface{}) error
	GetNull(context.Context) bool
	GetValue(context.Context) interface{}
}

// NewNullable creates a zero value of `target` (or the concrete type it's
// referencing, if it's a pointer) and calls its SetNull method.
//
// It is meant to be called through Into, not directly.
func NewNullable(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	receiver := pointerSafeZeroValue(ctx, target)
	method := receiver.MethodByName("SetNull")
	if !method.IsValid() {
		return target, path.NewErrorf("unexpectedly couldn't find SetUnknown method on type %s", receiver.Type().String())
	}
	results := method.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(val.IsNull()),
	})
	err := results[0].Interface()
	if err != nil {
		return target, path.NewError(err.(error))
	}
	return receiver, nil
}

// FromNullable creates an attr.Value from the data in a Nullable.
//
// It is meant to be called through OutOf, not directly.
func FromNullable(ctx context.Context, typ attr.Type, val Nullable, path *tftypes.AttributePath) (attr.Value, error) {
	if val.GetNull(ctx) {
		res, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), nil))
		if err != nil {
			return nil, path.NewError(err)
		}
		return res, nil
	}
	err := tftypes.ValidateValue(typ.TerraformType(ctx), val.GetValue(ctx))
	if err != nil {
		return nil, path.NewError(err)
	}
	res, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), val.GetValue(ctx)))
	if err != nil {
		return nil, path.NewError(err)
	}
	return res, nil
}

// NewValueConverter creates a zero value of `target` (or the concrete type
// it's referencing, if it's a pointer) and calls its FromTerraform5Value
// method.
//
// It is meant to be called through Into, not directly.
func NewValueConverter(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	receiver := pointerSafeZeroValue(ctx, target)
	method := receiver.MethodByName("FromTerraform5Value")
	if !method.IsValid() {
		return target, path.NewErrorf("unexpectedly couldn't find FromTerraform5Type method on type %s", receiver.Type().String())
	}
	results := method.Call([]reflect.Value{reflect.ValueOf(val)})
	err := results[0].Interface()
	if err != nil {
		return target, path.NewError(err.(error))
	}
	return receiver, nil
}

// FromValueCreator creates an attr.Value from the data in a
// tftypes.ValueCreator, calling its ToTerraform5Value method and converting
// the result to an attr.Value using `typ`.
//
// It is meant to be called from OutOf, not directly.
func FromValueCreator(ctx context.Context, typ attr.Type, val tftypes.ValueCreator, path *tftypes.AttributePath) (attr.Value, error) {
	raw, err := val.ToTerraform5Value()
	if err != nil {
		return nil, path.NewError(err)
	}
	err = tftypes.ValidateValue(typ.TerraformType(ctx), raw)
	if err != nil {
		return nil, path.NewError(err)
	}
	tfVal := tftypes.NewValue(typ.TerraformType(ctx), raw)
	res, err := typ.ValueFromTerraform(ctx, tfVal)
	if err != nil {
		return nil, path.NewError(err)
	}
	return res, nil
}

// NewAttributeValue creates a new reflect.Value by calling the
// ValueFromTerraform method on `typ`. It will return an error if the returned
// `attr.Value` is not the same type as `target`.
//
// It is meant to be called through Into, not directly.
func NewAttributeValue(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	res, err := typ.ValueFromTerraform(ctx, val)
	if err != nil {
		return target, err
	}
	if reflect.TypeOf(res) != target.Type() {
		return target, path.NewErrorf("can't use attr.Value %s, only %s is supported because %T is the type in the schema", target.Type(), reflect.TypeOf(res), typ)
	}
	return reflect.ValueOf(res), nil
}

// FromAttributeValue creates an attr.Value from an attr.Value. It just returns
// the attr.Value it is passed, but reserves the right in the future to do some
// validation on that attr.Value to make sure it matches the type produced by
// `typ`.
//
// It is meant to be called through OutOf, not directly.
func FromAttributeValue(ctx context.Context, typ attr.Type, val attr.Value, path *tftypes.AttributePath) (attr.Value, error) {
	return val, nil
}
