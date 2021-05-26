package reflect

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectPrimitive(ctx context.Context, val tftypes.Value, target interface{}, path *tftypes.AttributePath) error {
	realValue := trueReflectValue(target).Addr()
	err := val.As(&realValue)
	if err != nil {
		return path.NewError(err)
	}
	return nil
}
