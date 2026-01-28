// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package statestore

import "github.com/hashicorp/terraform-plugin-framework/diag"

// TODO: package docs
type UnlockRequest struct {
	StateID string
	LockID  string
}

// TODO: package docs
type UnlockResponse struct {
	Diagnostics diag.Diagnostics
}
