package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// build a string or bool depending on the type of `target`, and populate it
// with the data in `val`.
func reflectPrimitive(ctx context.Context, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) (reflect.Value, error) {
	switch target.Kind() {
	case reflect.Bool:
		var b bool
		err := val.As(&b)
		if err != nil {
			return target, path.NewError(err)
		}
		return reflect.ValueOf(b).Convert(target.Type()), nil
	case reflect.String:
		var s string
		err := val.As(&s)
		if err != nil {
			return target, path.NewError(err)
		}
		return reflect.ValueOf(s).Convert(target.Type()), nil
	default:
		return target, path.NewErrorf("unrecognized type %s (this should never happen)", target.Kind())
	}
}
