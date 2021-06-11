package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Pointer builds a new zero value of the concrete type that `target`
// references, populates it with BuildValue, and takes a pointer to it.
//
// It is meant to be called through Into, not directly.
func Pointer(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	if target.Kind() != reflect.Ptr {
		return target, path.NewErrorf("can't dereference pointer, not a pointer, is a %s (%s)", target.Type(), target.Kind())
	}
	// we may have gotten a nil pointer, so we need to create our own that
	// we can set
	pointer := reflect.New(target.Type().Elem())
	// build out whatever the pointer is pointing to
	pointed, err := BuildValue(ctx, typ, val, pointer.Elem(), opts, path)
	if err != nil {
		return target, err
	}
	// to be able to set the pointer to our new pointer, we need to create
	// a pointer to the pointer
	pointerPointer := reflect.New(pointer.Type())
	// we set the pointer we created on the pointer to the pointer
	pointerPointer.Elem().Set(pointer)
	// then it's settable, so we can now set the concrete value we created
	// on the pointer
	pointerPointer.Elem().Elem().Set(pointed)
	// return the pointer we created
	return pointerPointer.Elem(), nil
}

// create a zero value of concrete type underlying any number of pointers, then
// wrap it in that number of pointers again. The end result is to wind up with
// the same exact type, except now you can be sure it's pointing to actual data
// and will not give you a nil pointer dereference panic unexpectedly.
func pointerSafeZeroValue(ctx context.Context, target reflect.Value) reflect.Value {
	pointer := target.Type()
	var pointers int
	for pointer.Kind() == reflect.Ptr {
		pointer = pointer.Elem()
		pointers++
	}
	receiver := reflect.Zero(pointer)
	for i := 0; i < pointers; i++ {
		newReceiver := reflect.New(receiver.Type())
		newReceiver.Elem().Set(receiver)
		receiver = newReceiver
	}
	return receiver
}

func FromPointer(ctx context.Context, typ attr.Type, value reflect.Value, path *tftypes.AttributePath) (attr.Value, error) {
	if value.Kind() != reflect.Ptr {
		return nil, path.NewErrorf("can't use type %s as a pointer", value.Type())
	}
	if value.IsNil() {
		return typ.ValueFromTerraform(ctx, tftypes.NewValue(typ.TerraformType(ctx), nil))
	}
	return FromValue(ctx, typ, value.Elem().Interface(), path)
}
