// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

// GetStatesRequest represents a request for the StateStore to return metadata,
// such as its type name. An instance of this request struct is supplied as
// an argument to the StateStore type Get method.
type GetStatesRequest struct {
	// ProviderTypeName is the string returned from
	// [provider.GetStatesResponse.TypeName], if the Provider type implements
	// the Get method. This string should prefix the StateStore type name
	// with an underscore in the response.
	ProviderTypeName string
}

// GetStatesResponse represents a response to a GetStatesRequest. An
// instance of this response struct is supplied as an argument to the
// StateStore type Get method.
type GetStatesResponse struct {
	// TypeName should be the full state store type, including the provider
	// type prefix and an underscore. For example, examplecloud_thing.
	TypeName string
}
