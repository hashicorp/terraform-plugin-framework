package resource

// TypeNameRequest represents a request for the Resource to return its type
// name. An instance of this request struct is supplied as an argument to the
// Resource type TypeName method.
type TypeNameRequest struct {
	// ProviderTypeName is the string returned from
	// [provider.MetadataResponse.TypeName], if the Provider type implements
	// the Metadata method. This string should prefix the Resource type name
	// with an underscore in the response.
	ProviderTypeName string
}

// TypeNameResponse represents a response to a TypeNameRequest. An
// instance of this response struct is supplied as an argument to the
// Resource type TypeName method.
type TypeNameResponse struct {
	// TypeName should be the full resource type, including the provider
	// type prefix and an underscore. For example, examplecloud_thing.
	TypeName string
}
