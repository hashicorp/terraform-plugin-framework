package attrpath

import "fmt"

// Error represents an error associated with part of an attr.Value or
// tfsdk.Schema, indicated by the Path property.
type Error struct {
	Path Path
	err  error
}

// Equal returns true if two Errors are semantically equal. To be considered
// equal, they must have the same path and if errors are set, the strings
// returned by their `Error()` methods must match.
func (a Error) Equal(o Error) bool {
	if !a.Path.Equal(o.Path) {
		return false
	}

	if (a.err == nil && o.err != nil) || (a.err != nil && o.err == nil) {
		return false
	}

	if a.err == nil {
		return true
	}

	return a.err.Error() == o.err.Error()
}

func (a Error) Error() string {
	var path string
	if !a.Path.IsEmpty() {
		path = a.Path.String() + ": "
	}
	return fmt.Sprintf("%s%s", path, a.err)
}

func (a Error) Unwrap() error {
	return a.err
}
