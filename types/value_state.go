package types

import "fmt"

const (
	// valueStateDeprecated represents a value where it can potentially be
	// controlled by exported fields such as Null and Unknown. Since consumers
	// can adjust those fields directly and it would not modify the internal
	// valueState value, this sentinel value is a placeholder which can be
	// checked in logic before assuming the valueState value is accurate.
	//
	// This value is 0 so it is the zero-value of existing implementations to
	// preserve existing behaviors. A future version will switch the zero-value
	// to null and export this implementation in the attr package.
	valueStateDeprecated valueState = 0

	// valueStateNull represents a value which is null.
	valueStateNull valueState = 1

	// valueStateUnknown represents a value which is unknown.
	valueStateUnknown valueState = 2

	// valueStateKnown represents a value which is known (not null or unknown).
	valueStateKnown valueState = 3
)

type valueState uint8

func (s valueState) String() string {
	switch s {
	case valueStateDeprecated:
		return "deprecated"
	case valueStateKnown:
		return "known"
	case valueStateNull:
		return "null"
	case valueStateUnknown:
		return "unknown"
	default:
		panic(fmt.Sprintf("unhandled valueState in String: %d", s))
	}
}
