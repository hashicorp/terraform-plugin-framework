// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CancelRequest represents a request for the provider to cancel a
// resource. An instance of this request struct is supplied as an argument to
// the resource's Cancel function.
type CancelRequest struct {
	CancellationToken string
	CancellationType  CancelType
}

// CancelResponse represents a response to a CancelRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Cancel function, in which the provider
// should set values on the CancelResponse as appropriate.
type CancelResponse struct {
	Diagnostics diag.Diagnostics
}

const (
	ActionCancelTypeSoft CancelType = 0
	ActionCancelTypeHard CancelType = 1
)

type ActionCancelType struct {
	CancelType CancelType
}

type CancelType int32

func (c CancelType) String() string {
	switch c {
	case 0:
		return "SOFT"
	case 1:
		return "HARD"
	}
	return "UNKNOWN"
}
