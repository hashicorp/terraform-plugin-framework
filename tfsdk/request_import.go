package tfsdk

// ImportResourceStateRequest represents a request for the provider to import a
// resource. An instance of this request struct is supplied as an argument to
// the Resource's ImportState method.
type ImportResourceStateRequest struct {
	// ID represents the import identifier supplied by the practitioner when
	// calling the import command. In many cases, this may align with the
	// unique identifier for the resource, which can optionally be stored
	// as an Attribute. However, this identifier can also be treated as
	// its own type of value and parsed during import. This value
	// is not stored in the state unless the provider explicitly stores it.
	ID string
}
