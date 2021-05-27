package reflect

import (
	"context"
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectNumber(ctx context.Context, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) error {
	realValue := trueReflectValue(target)
	if !realValue.CanSet() {
		return path.NewErrorf("can't set %s", realValue.Type())
	}
	result := big.NewFloat(0)
	err := val.As(&result)
	if err != nil {
		return path.NewError(err)
	}
	roundingError := path.NewErrorf("can't store %s in %s", result.String(), target.Type())
	switch realValue.Type() {
	case reflect.TypeOf(big.NewFloat(0)):
		realValue.Set(reflect.ValueOf(result))
		return nil
	case reflect.TypeOf(big.NewInt(0)):
		intResult, acc := result.Int(nil)
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return roundingError
		}
		realValue.Set(reflect.ValueOf(intResult))
		return nil
	}
	switch realValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		intResult, acc := result.Int64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return roundingError
		}
		switch realValue.Kind() {
		case reflect.Int:
			if strconv.IntSize == 32 && intResult > math.MaxInt32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MaxInt32
			}
			if strconv.IntSize == 32 && intResult < math.MinInt32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MinInt32
			}
			realValue.Set(reflect.ValueOf(int(intResult)))
		case reflect.Int8:
			if intResult > math.MaxInt8 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MaxInt8
			}
			if intResult < math.MinInt8 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MinInt8
			}
			realValue.Set(reflect.ValueOf(int8(intResult)))
		case reflect.Int16:
			if intResult > math.MaxInt16 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MaxInt16
			}
			if intResult < math.MinInt16 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MinInt16
			}
			realValue.Set(reflect.ValueOf(int16(intResult)))
		case reflect.Int32:
			if intResult > math.MaxInt32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MaxInt32
			}
			if intResult < math.MinInt32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				intResult = math.MinInt32
			}
			realValue.Set(reflect.ValueOf(int32(intResult)))
		case reflect.Int64:
			realValue.Set(reflect.ValueOf(intResult))
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		uintResult, acc := result.Uint64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return roundingError
		}
		switch realValue.Kind() {
		case reflect.Uint:
			if strconv.IntSize == 32 && uintResult > math.MaxUint32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				uintResult = math.MaxUint32
			}
			realValue.Set(reflect.ValueOf(uint(uintResult)))
		case reflect.Uint8:
			if uintResult > math.MaxUint8 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				uintResult = math.MaxUint8
			}
			realValue.Set(reflect.ValueOf(uint8(uintResult)))
		case reflect.Uint16:
			if uintResult > math.MaxUint16 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				uintResult = math.MaxUint16
			}
			realValue.Set(reflect.ValueOf(uint16(uintResult)))
		case reflect.Uint32:
			if uintResult > math.MaxUint32 {
				if !opts.AllowRoundingNumbers {
					return roundingError
				}
				uintResult = math.MaxUint32
			}
			realValue.Set(reflect.ValueOf(uint32(uintResult)))
		case reflect.Uint64:
			realValue.Set(reflect.ValueOf(uintResult))
		}
		return nil
	case reflect.Float32:
		floatResult, acc := result.Float32()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return roundingError
		} else if acc == big.Above {
			floatResult = math.MaxFloat32
		} else if acc == big.Below {
			floatResult = math.SmallestNonzeroFloat32
		} else if acc != big.Exact {
			return path.NewErrorf("unsure how to round %s and %f", acc, floatResult)
		}
		realValue.Set(reflect.ValueOf(floatResult))
		return nil
	case reflect.Float64:
		floatResult, acc := result.Float64()
		if acc != big.Exact && !opts.AllowRoundingNumbers {
			return roundingError
		}
		if acc == big.Above {
			if floatResult == math.Inf(1) || floatResult == math.MaxFloat64 {
				floatResult = math.MaxFloat64
			} else if floatResult == 0.0 || floatResult == math.SmallestNonzeroFloat64 {
				floatResult = -math.SmallestNonzeroFloat64
			} else {
				return path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
			}
		} else if acc == big.Below {
			if floatResult == math.Inf(-1) || floatResult == -math.MaxFloat64 {
				floatResult = -math.MaxFloat64
			} else if floatResult == -0.0 || floatResult == -math.SmallestNonzeroFloat64 {
				floatResult = math.SmallestNonzeroFloat64
			} else {
				return path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
			}
		} else if acc != big.Exact {
			return path.NewErrorf("not sure how to round %s and %f", acc, floatResult)
		}
		realValue.Set(reflect.ValueOf(floatResult))
		return nil
	}
	return path.NewErrorf("can't set number on %s", realValue.Type())
}
