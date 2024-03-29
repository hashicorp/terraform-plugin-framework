package function

import (
	"github.com/hashicorp/terraform-plugin-framework/function/validator"
)

// ParameterWithBoolValidators is an optional interface on Parameter which
// enables Bool validation support.
type ParameterWithBoolValidators interface {
	Parameter

	// BoolValidators should return a list of Bool validators.
	BoolValidators() []validator.Bool
}

// ParameterWithInt64Validators is an optional interface on Parameter which
// enables Int64 validation support.
type ParameterWithInt64Validators interface {
	Parameter

	// Int64Validators should return a list of Int64 validators.
	Int64Validators() []validator.Int64
}

// ParameterWithFloat64Validators is an optional interface on Parameter which
// enables Float64 validation support.
type ParameterWithFloat64Validators interface {
	Parameter

	// Float64Validators should return a list of Float64 validators.
	Float64Validators() []validator.Float64
}

// ParameterWithDynamicValidators is an optional interface on Parameter which
// enables Dynamic validation support.
type ParameterWithDynamicValidators interface {
	Parameter

	// DynamicValidators should return a list of Dynamic validators.
	DynamicValidators() []validator.Dynamic
}

// ParameterWithListValidators is an optional interface on Parameter which
// enables List validation support.
type ParameterWithListValidators interface {
	Parameter

	// ListValidators should return a list of List validators.
	ListValidators() []validator.List
}

// ParameterWithMapValidators is an optional interface on Parameter which
// enables Map validation support.
type ParameterWithMapValidators interface {
	Parameter

	// MapValidators should return a list of Map validators.
	MapValidators() []validator.Map
}

// ParameterWithNumberValidators is an optional interface on Parameter which
// enables Number validation support.
type ParameterWithNumberValidators interface {
	Parameter

	// NumberValidators should return a list of Map validators.
	NumberValidators() []validator.Number
}

// ParameterWithObjectValidators is an optional interface on Parameter which
// enables Object validation support.
type ParameterWithObjectValidators interface {
	Parameter

	// ObjectValidators should return a list of Object validators.
	ObjectValidators() []validator.Object
}

// ParameterWithSetValidators is an optional interface on Parameter which
// enables Set validation support.
type ParameterWithSetValidators interface {
	Parameter

	// SetValidators should return a list of Set validators.
	SetValidators() []validator.Set
}

// ParameterWithStringValidators is an optional interface on Parameter which
// enables String validation support.
type ParameterWithStringValidators interface {
	Parameter

	// StringValidators should return a list of String validators.
	StringValidators() []validator.String
}
