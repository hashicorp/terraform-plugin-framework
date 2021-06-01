package attr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Value defines an interface for describing data associated with an attribute.
// Values allow provider developers to specify data in a convenient format, and
// have it transparently be converted to formats Terraform understands.
type Value interface {
	// ToTerraformValue returns the data contained in the Value as
	// a Go type that tftypes.NewValue will accept.
	ToTerraformValue(context.Context) (interface{}, error)

	// SetTerraformValue updates the data in Value to match the
	// passed tftypes.Value.
	SetTerraformValue(context.Context, tftypes.Value) error

	// Equal must return true if the Value is considered semantically equal
	// to the Value passed as an argument.
	Equal(Value) bool
}
