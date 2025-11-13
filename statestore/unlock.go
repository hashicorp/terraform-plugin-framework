// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import "github.com/hashicorp/terraform-plugin-framework/diag"

// UnlockStateRequest represents a request for the StateStore to return metadata,
// such as its type name. An instance of this request struct is supplied as
// an argument to the StateStore type Unlock method.
type UnlockStateRequest struct {
	// TypeName should be the full state store type, including the provider
	// type prefix and an underscore. For example, examplecloud_thing.
	TypeName string
	StateId  string
	LockId   string
}

// UnlockStateResponse represents a response to a UnlockStateRequest. An
// instance of this response struct is supplied as an argument to the
// StateStore type Unlock method.
type UnlockStateResponse struct {
	Diagnostics []*diag.Diagnostic
}
