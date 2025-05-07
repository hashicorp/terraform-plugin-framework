// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

type MetadataRequest struct {
	ProviderTypeName string
}

type MetadataResponse struct {
	TypeName string
}
