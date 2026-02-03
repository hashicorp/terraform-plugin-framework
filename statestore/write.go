// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// TODO: package docs
type WriteRequest struct {
	StateID    string
	StateBytes []byte
}

// TODO: package docs
type WriteResponse struct {
	Diagnostics diag.Diagnostics
}
