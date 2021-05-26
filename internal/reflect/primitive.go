package reflect

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPrimitive(ctx context.Context, val tftypes.Value, target interface{}, path *tftypes.AttributePath) error {
	realValue := trueReflectValue(target)
	if !realValue.CanAddr() {
		return path.NewErrorf("can't obtain address of %T", target)
	}
	err := val.As(realValue.Addr())
	if err != nil {
		return path.NewError(err)
	}
	return nil
}
