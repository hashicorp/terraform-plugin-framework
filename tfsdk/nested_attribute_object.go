package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwschema.NestedAttributeObject = nestedAttributeObject{}

// nestedAttributeObject is the object containing the underlying attributes
// for an Attribute with nested attributes. This is a temporary type until
// Attribute is removed.
type nestedAttributeObject struct {
	// Attributes is the mapping of underlying attribute names to attribute
	// definitions. This field must be set.
	Attributes map[string]fwschema.Attribute
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o nestedAttributeObject) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.NestedAttributeObjectApplyTerraform5AttributePathStep(o, step)
}

// Equal returns true if the given nestedAttributeObject is equivalent.
func (o nestedAttributeObject) Equal(other fwschema.NestedAttributeObject) bool {
	if _, ok := other.(nestedAttributeObject); !ok {
		return false
	}

	return fwschema.NestedAttributeObjectEqual(o, other)
}

// GetAttributes returns the Attributes field value.
func (o nestedAttributeObject) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// Type returns the framework type of the nestedAttributeObject.
func (o nestedAttributeObject) Type() types.ObjectTypable {
	return fwschema.NestedAttributeObjectType(o)
}
