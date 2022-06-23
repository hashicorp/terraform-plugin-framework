package path

import "strings"

// Expressions is a collection of attribute path expressions.
type Expressions []Expression

// String returns the human-readable representation of the expression
// collection. It is intended for logging and error messages and is not
// protected by compatibility guarantees.
//
// Empty expressions are skipped.
func (p Expressions) String() string {
	var result strings.Builder

	result.WriteString("[")

	for pathIndex, path := range p {
		if path.Equal(Expression{}) {
			continue
		}

		if pathIndex != 0 {
			result.WriteString(",")
		}

		result.WriteString(path.String())
	}

	result.WriteString("]")

	return result.String()
}
