package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPointer(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	if target.Kind() != reflect.Ptr {
		return target, path.NewErrorf("can't dereference pointer, not a pointer, is a %s (%s)", target.Type(), target.Kind())
	}
	// we may have gotten a nil pointer, so we need to create our own that
	// we can set
	pointer := reflect.New(target.Type().Elem())
	// build out whatever the pointer is pointing to
	pointed, err := buildReflectValue(ctx, val, pointer.Elem(), opts, path)
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

func pointerSafeZeroValue(ctx context.Context, target reflect.Value) reflect.Value {
	pointer := target.Type()
	var pointers int
	for pointer.Kind() == reflect.Ptr {
		pointer = pointer.Elem()
		pointers += 1
	}
	receiver := reflect.Zero(pointer)
	for i := 0; i < pointers; i++ {
		newReceiver := reflect.New(receiver.Type())
		newReceiver.Elem().Set(receiver)
		receiver = newReceiver
	}
	return receiver
}
