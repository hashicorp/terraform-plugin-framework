// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package statestore

import "github.com/hashicorp/terraform-plugin-framework/diag"

// TODO: package docs
type LockRequest struct {
	StateID   string
	Operation string
}

// TODO: package docs
type LockResponse struct {
	LockID      string
	Diagnostics diag.Diagnostics
}
