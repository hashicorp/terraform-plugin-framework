package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPrimitive(ctx context.Context, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) (reflect.Value, error) {
	switch target.Kind() {
	case reflect.Bool:
		var b bool
		err := val.As(&b)
		if err != nil {
			return target, path.NewError(err)
		}
		return reflect.ValueOf(b), nil
	case reflect.String:
		var s string
		err := val.As(&s)
		if err != nil {
			return target, path.NewError(err)
		}
		return reflect.ValueOf(s), nil
	default:
		return target, path.NewErrorf("unrecognized type %s (this should never happen)", target.Kind())
	}
}
