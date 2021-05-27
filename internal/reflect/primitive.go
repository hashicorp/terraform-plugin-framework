package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPrimitive(ctx context.Context, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) error {
	realValue := trueReflectValue(target)
	if !realValue.CanAddr() {
		return path.NewErrorf("can't obtain address of %s", target.Type())
	}
	err := val.As(realValue.Addr().Interface())
	if err != nil {
		return path.NewError(err)
	}
	return nil
}
