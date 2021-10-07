package types

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func numberValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Number{Unknown: true}, nil
	}
	if in.IsNull() {
		return Number{Null: true}, nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return Number{Value: n}, nil
}

var _ attr.Value = Number{}

// Number represents a number value, exposed as a *big.Float. Numbers can be
// floats or integers.
type Number struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value *big.Float
}

// Type returns a NumberType.
func (n Number) Type(_ context.Context) attr.Type {
	return NumberType
}

// ToTerraformValue returns the data contained in the *Number as a *big.Float.
// If Unknown is true, it returns a tftypes.UnknownValue. If Null is true, it
// returns nil.
func (n Number) ToTerraformValue(_ context.Context) (interface{}, error) {
	if n.Null {
		return nil, nil
	}
	if n.Unknown {
		return tftypes.UnknownValue, nil
	}
	if n.Value == nil {
		return nil, nil
	}
	return n.Value, nil
}

// Equal returns true if `other` is a *Number and has the same value as `n`.
func (n Number) Equal(other attr.Value) bool {
	o, ok := other.(Number)
	if !ok {
		return false
	}
	if n.Unknown != o.Unknown {
		return false
	}
	if n.Null != o.Null {
		return false
	}
	if n.Value == nil && o.Value == nil {
		return true
	}
	if n.Value == nil || o.Value == nil {
		return false
	}
	return n.Value.Cmp(o.Value) == 0
}

func (n Number) MarshalJSON() ([]byte, error) {
	if n.Null || n.Unknown {
		return json.Marshal((*big.Float)(nil))
	}
	// big.Float implements the text marshaler which will wrap the result
	// in double quotes which is incorrect for JSON.
	// The docs state it only marshals the float anyways so using the
	// float64 primitive json marshaler is hopefully good enough.
	f, _ := n.Value.Float64()
	return json.Marshal(f)
}

func (n *Number) UnmarshalJSON(data []byte) error {
	var fPtr *float64
	if err := json.Unmarshal(data, &fPtr); err != nil {
		return err
	}
	n.Unknown = false
	if fPtr == nil {
		n.Value = nil
		n.Null = true
	} else {
		n.Value = big.NewFloat(*fPtr)
		n.Null = false
	}
	return nil
}
