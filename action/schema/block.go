// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

// Block defines a structural field inside an action type schema. Implementations in this
// package include:
//   - ListNestedBlock
//   - SetNestedBlock
//   - SingleNestedBlock
//
// In practitioner configurations, an equals sign (=) cannot be used to set the
// value. Blocks are instead repeated as necessary, or require the use of
// [Dynamic Block Expressions].
//
// Prefer NestedAttribute over Block. Blocks should typically be used for
// configuration compatibility with previously existing schemas from an older
// Terraform Plugin SDK. Efforts should be made to convert from Block to
// NestedAttribute as a breaking change for practitioners.
//
// [Dynamic Block Expressions]: https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks
//
// [Configuration Reference]: https://developer.hashicorp.com/terraform/language/syntax/configuration
type Block interface {
	fwschema.Block
}

// schemaBlocks is an action block to fwschema type conversion function.
func schemaBlocks(blocks map[string]Block) map[string]fwschema.Block {
	result := make(map[string]fwschema.Block, len(blocks))

	for name, block := range blocks {
		result[name] = block
	}

	return result
}
