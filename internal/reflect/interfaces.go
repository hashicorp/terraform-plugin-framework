package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type setUnknownable interface {
	SetUnknown(context.Context, bool) error
}

func reflectUnknownable(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

type setNullable interface {
	SetNull(context.Context, bool) error
}

func reflectNullable(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

func reflectValueConverter(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
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

func reflectAttributeValue(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	receiver := pointerSafeZeroValue(ctx, target)
	method := receiver.MethodByName("SetTerraformValue")
	if !method.IsValid() {
		return target, path.NewErrorf("unexpectedly couldn't find SetTeraformValue method on type %s", receiver.Type().String())
	}
	results := method.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(val),
	})
	err := results[0].Interface()
	if err != nil {
		return target, path.NewError(err.(error))
	}
	return receiver, nil
}
