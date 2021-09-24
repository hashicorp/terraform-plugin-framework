package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// transformFunc is the signature expected for tftypes.Transform functions.
type transformFunc func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error)
