// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package validator

// ValidateSchemaClientCapabilities allows Terraform to publish information
// regarding optionally supported protocol features for the schema validation
// RPCs, such as forward-compatible Terraform behavior changes.
type ValidateSchemaClientCapabilities struct {
	// WriteOnlyAttributesAllowed indicates that the Terraform client
	// initiating the request supports write-only attributes for managed
	// resources.
	//
	// This client capability is only populated during managed resource schema
	// validation.
	WriteOnlyAttributesAllowed bool

	// ComputedBlocksAllowed indicates that the Terraform client
	// initiating the request supports computed blocks for managed
	// resources.
	//
	// This client capability is only populated during managed resource schema
	// validation and only for protocol version 6, as computed blocks are not
	// supported by protocol version 5.
	ComputedBlocksAllowed bool
}
