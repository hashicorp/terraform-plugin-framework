package reflect

import (
	"context"
	"math/big"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// OutOf is the inverse of Into, taking a Go value (val) and transforming it
// into an (attr.Value, attr.Type) pair. Each Go type present in val must have
// an appropriate attr.Type supplied via opts.
func OutOf(ctx context.Context, typ attr.Type, val interface{}, opts OutOfOptions) (attr.Value, error) {
	return FromValue(ctx, typ, val, opts, tftypes.NewAttributePath())
}

func FromValue(ctx context.Context, typ attr.Type, val interface{}, opts OutOfOptions, path *tftypes.AttributePath) (attr.Value, error) {
	if bf, ok := val.(*big.Float); ok {
		return FromBigFloat(ctx, typ, bf, opts, path)
	} else if bi, ok := val.(*big.Int); ok {
		return FromBigInt(ctx, typ, bi, opts, path)
	}
	value := reflect.ValueOf(val)
	kind := value.Kind()
	switch kind {
	case reflect.Struct:
		return FromStruct(ctx, typ.(attr.TypeWithAttributeTypes), value, opts, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return FromInt(ctx, typ, value.Int(), opts, path)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return FromUint(ctx, typ, value.Uint(), opts, path)
	case reflect.Float32, reflect.Float64:
		return FromFloat(ctx, typ, value.Float(), opts, path)
	case reflect.Bool:
		return FromBool(ctx, typ, value.Bool(), opts, path)
	case reflect.String:
		return FromString(ctx, typ, value.String(), opts, path)
	case reflect.Slice:
		return FromSlice(ctx, typ, value, opts, path)
	case reflect.Map:
		return FromMap(ctx, typ.(attr.TypeWithElementType), value, opts, path)
	case reflect.Ptr:
		return FromPointer(ctx, typ, value, opts, path)
	default:
		return nil, path.NewErrorf("don't know how to construct attr.Type from %T (%s)", val, kind)
	}
}
