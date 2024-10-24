package refinement

import "fmt"

type Key int64

func (k Key) String() string {
	// TODO: Not sure when this is used, double check the names
	switch k {
	// TODO: is this the right name for it?
	case KeyNotNull:
		return "not_null"
	case KeyStringPrefix:
		return "string_prefix"
	default:
		return fmt.Sprintf("unsupported refinement: %d", k)
	}
}

const (
	// MAINTAINER NOTE: This is named slightly different from the terraform-plugin-go `Nullness` refinement it maps to.
	// This is done because framework only support nullness refinements that indicate an unknown value is definitely not null.
	// Values that are definitely null should be represented as a known null value instead.
	KeyNotNull      = Key(1)
	KeyStringPrefix = Key(2)
	// KeyNumberLowerBound           = Key(3)
	// KeyNumberUpperBound           = Key(4)
	// KeyCollectionLengthLowerBound = Key(5)
	// KeyCollectionLengthUpperBound = Key(6)
)

type Refinement interface {
	Equal(Refinement) bool
	String() string
	unimplementable() // prevents external implementations, all refinements are defined in the Terraform/HCL type system go-cty.
}

type Refinements map[Key]Refinement

func (r Refinements) Equal(o Refinements) bool {
	return false
}
func (r Refinements) String() string {
	// TODO: Not sure when this is used, should just aggregate and call all underlying refinements.String() method
	return "todo"
}
