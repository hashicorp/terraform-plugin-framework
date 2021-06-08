package reflect

import (
	"context"
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Number creates a *big.Float and populates it with the data in `val`. It then
// gets converted to the type of `target`, as long as `target` is a valid
// number type (any of the built-in int, uint, or float types, *big.Float, and
// *big.Int).
//
// Number will loudly fail when a number cannot be losslessly represented using
// the requested type, unless opts.AllowRoundingNumbers is set to true. This
// setting is mildly dangerous, because Terraform does not like when you round
// things, as a general rule of thumb.
//
// It is meant to be called through Into, not directly.
func Number(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	result := big.NewFloat(0)
	err := val.As(&result)
	if err != nil {
		return target, path.NewError(err)
	}
	roundingError := path.NewErrorf("can't store %s in %s", result.String(), target.Type())
	switch target.Type() {
	case reflect.TypeOf(big.NewFloat(0)):
		return reflect.ValueOf(result), nil
	case reflect.TypeOf(big.NewInt(0)):
		intResult, acc := result.Int(nil)
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return reflect.ValueOf(result), roundingError
		}
		return reflect.ValueOf(intResult), nil
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		intResult, acc := result.Int64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return target, roundingError
		}
		switch target.Kind() {
		case reflect.Int:
			if strconv.IntSize == 32 && intResult > math.MaxInt32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MaxInt32
			}
			if strconv.IntSize == 32 && intResult < math.MinInt32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MinInt32
			}
			return reflect.ValueOf(int(intResult)), nil
		case reflect.Int8:
			if intResult > math.MaxInt8 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MaxInt8
			}
			if intResult < math.MinInt8 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MinInt8
			}
			return reflect.ValueOf(int8(intResult)), nil
		case reflect.Int16:
			if intResult > math.MaxInt16 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MaxInt16
			}
			if intResult < math.MinInt16 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MinInt16
			}
			return reflect.ValueOf(int16(intResult)), nil
		case reflect.Int32:
			if intResult > math.MaxInt32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MaxInt32
			}
			if intResult < math.MinInt32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				intResult = math.MinInt32
			}
			return reflect.ValueOf(int32(intResult)), nil
		case reflect.Int64:
			return reflect.ValueOf(intResult), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		uintResult, acc := result.Uint64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return target, roundingError
		}
		switch target.Kind() {
		case reflect.Uint:
			if strconv.IntSize == 32 && uintResult > math.MaxUint32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				uintResult = math.MaxUint32
			}
			return reflect.ValueOf(uint(uintResult)), nil
		case reflect.Uint8:
			if uintResult > math.MaxUint8 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				uintResult = math.MaxUint8
			}
			return reflect.ValueOf(uint8(uintResult)), nil
		case reflect.Uint16:
			if uintResult > math.MaxUint16 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				uintResult = math.MaxUint16
			}
			return reflect.ValueOf(uint16(uintResult)), nil
		case reflect.Uint32:
			if uintResult > math.MaxUint32 {
				if !opts.AllowRoundingNumbers {
					return target, roundingError
				}
				uintResult = math.MaxUint32
			}
			return reflect.ValueOf(uint32(uintResult)), nil
		case reflect.Uint64:
			return reflect.ValueOf(uintResult), nil
		}
	case reflect.Float32:
		floatResult, acc := result.Float32()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return target, roundingError
		} else if acc == big.Above {
			floatResult = math.MaxFloat32
		} else if acc == big.Below {
			floatResult = math.SmallestNonzeroFloat32
		} else if acc != big.Exact {
			return target, path.NewErrorf("unsure how to round %s and %f", acc, floatResult)
		}
		return reflect.ValueOf(floatResult), nil
	case reflect.Float64:
		floatResult, acc := result.Float64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return target, roundingError
		}
		if acc == big.Above {
			if floatResult == math.Inf(1) || floatResult == math.MaxFloat64 {
				floatResult = math.MaxFloat64
			} else if floatResult == 0.0 || floatResult == math.SmallestNonzeroFloat64 {
				floatResult = -math.SmallestNonzeroFloat64
			} else {
				return target, path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
			}
		} else if acc == big.Below {
			if floatResult == math.Inf(-1) || floatResult == -math.MaxFloat64 {
				floatResult = -math.MaxFloat64
			} else if floatResult == -0.0 || floatResult == -math.SmallestNonzeroFloat64 {
				floatResult = math.SmallestNonzeroFloat64
			} else {
				return target, path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
			}
		} else if acc != big.Exact {
			return target, path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
		}
		return reflect.ValueOf(floatResult), nil
	}
	return target, path.NewErrorf("can't convert number to %s", target.Type())
}
