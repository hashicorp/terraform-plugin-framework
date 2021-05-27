package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPrimitive(ctx context.Context, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) error {
	realValue := trueReflectValue(target)
	if !realValue.CanSet() {
		return path.NewErrorf("can't set %s", target.Type())
	}
	switch realValue.Kind() {
	case reflect.Bool:
		var b bool
		err := val.As(&b)
		if err != nil {
			return path.NewError(err)
		}
		realValue.SetBool(b)
	case reflect.String:
		var s string
		err := val.As(&s)
		if err != nil {
			return path.NewError(err)
		}
		realValue.SetString(s)
	}
	return nil
}
