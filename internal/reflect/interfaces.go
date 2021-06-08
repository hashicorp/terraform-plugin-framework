package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// SetUnknownable is an interface for types that can be explicitly set to known
// or unknown.
type SetUnknownable interface {
	SetUnknown(context.Context, bool) error
}

// Unknownable creates a zero value of `target` (or the concrete type it's
// referencing, if it's a pointer) and calls its SetUnknown method.
//
// It is meant to be called through Into, not directly.
func Unknownable(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

// SetNullable is an interface for types that can be explicitly set to null.
type SetNullable interface {
	SetNull(context.Context, bool) error
}

// Nullable creates a zero value of `target` (or the concrete type it's
// referencing, if it's a pointer) and calls its SetNull method.
//
// It is meant to be called through Into, not directly.
func Nullable(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

// ValueConverter creates a zero value of `target` (or the concrete type it's
// referencing, if it's a pointer) and calls its FromTerraform5Value method.
//
// It is meant to be called through Into, not directly.
func ValueConverter(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

// AttributeValue creates a new reflect.Value by calling the ValueFromTerraform
// method on `typ`. It will return an error if the returned `attr.Value` is not
// the same type as `target`.
//
// It is meant to be called through Into, not directly.
func AttributeValue(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	res, err := typ.ValueFromTerraform(ctx, val)
	if err != nil {
		return target, err
	}
	if reflect.TypeOf(res) != target.Type() {
		return target, path.NewErrorf("can't use attr.Value %s, only %s is supported because %T is the type in the schema", target.Type(), reflect.TypeOf(res), typ)
	}
	return reflect.ValueOf(res), nil
}
