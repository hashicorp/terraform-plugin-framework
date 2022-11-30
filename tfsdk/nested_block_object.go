package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwschema.NestedBlockObject = nestedBlockObject{}

// nestedBlockObject is the object containing the underlying attributes and
// blocks for Block. It is a temporary type until Block is removed.
//
// This object enables customizing and simplifying details within its parent
// Block, therefore it cannot have Terraform schema fields such as Description,
// etc.
type nestedBlockObject struct {
	// Attributes is the mapping of underlying attribute names to attribute
	// definitions.
	//
	// Names must only contain lowercase letters, numbers, and underscores.
	// Names must not collide with any Blocks names.
	Attributes map[string]Attribute

	// Blocks is the mapping of underlying block names to block definitions.
	//
	// Names must only contain lowercase letters, numbers, and underscores.
	// Names must not collide with any Attributes names.
	Blocks map[string]Block
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o nestedBlockObject) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.NestedBlockObjectApplyTerraform5AttributePathStep(o, step)
}

// Equal returns true if the given nestedBlockObject is equivalent.
func (o nestedBlockObject) Equal(other fwschema.NestedBlockObject) bool {
	if _, ok := other.(nestedBlockObject); !ok {
		return false
	}

	return fwschema.NestedBlockObjectEqual(o, other)
}

// GetAttributes returns the Attributes field value.
func (o nestedBlockObject) GetAttributes() fwschema.UnderlyingAttributes {
	return schemaAttributes(o.Attributes)
}

// GetAttributes returns the Blocks field value.
func (o nestedBlockObject) GetBlocks() map[string]fwschema.Block {
	return schemaBlocks(o.Blocks)
}

// Type returns the framework type of the nestedBlockObject.
func (o nestedBlockObject) Type() basetypes.ObjectTypable {
	return fwschema.NestedBlockObjectType(o)
}
