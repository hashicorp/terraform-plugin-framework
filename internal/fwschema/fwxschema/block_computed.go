// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwxschema

import "github.com/hashicorp/terraform-plugin-framework/internal/fwschema"

// BlockWithComputed is an optional interface on Block which enables
// protocol-level support for computed nested blocks.
type BlockWithComputed interface {
	fwschema.Block

	// GetComputed returns whether the block should be marked computed in
	// provider schema responses.
	GetComputed() bool
}
