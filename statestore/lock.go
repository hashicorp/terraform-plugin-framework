// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

// LockRequest represents a request for the StateStore to return metadata,
// such as its type name. An instance of this request struct is supplied as
// an argument to the StateStore type Lock method.
type LockRequest struct {
	// ProviderTypeName is the string returned from
	// [provider.LockResponse.TypeName], if the Provider type implements
	// the Lock method. This string should prefix the StateStore type name
	// with an underscore in the response.
	ProviderTypeName string
}

// LockResponse represents a response to a LockRequest. An
// instance of this response struct is supplied as an argument to the
// StateStore type Lock method.
type LockResponse struct {
	// TypeName should be the full state store type, including the provider
	// type prefix and an underscore. For example, examplecloud_thing.
	TypeName string
}
