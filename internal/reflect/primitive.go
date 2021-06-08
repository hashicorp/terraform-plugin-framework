package reflect

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Primitive builds a string or boolean, depending on the type of `target`, and
// populates it with the data in `val`.
//
// It is meant to be called through `Into`, not directly.
func Primitive(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, path *tftypes.AttributePath) (reflect.Value, error) {
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
