package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block defines the constraints and behaviors of a single structural field in a
// schema.
type Block struct {
	tftypes.AttributePathStepper

	// Attributes
	Attributes map[string]Attribute

	// Blocks can have their own nested blocks. This nested map of blocks
	// behaves exactly like the map of blocks on the Schema type.
	Blocks map[string]Block

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// NestingMode indicates the block kind.
	NestingMode NestingMode
}

// ApplyTerraform5AttributePathStep transparently calls
// ApplyTerraform5AttributePathStep on a.Type or a.Attributes,
// whichever is non-nil. It allows Attributes to be walked using tftypes.Walk
// and tftypes.Transform.
// func (b Block) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
// 	return b.NestingMode.ApplyTerraform5AttributePathStep(step)
// }
