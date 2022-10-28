package fwxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeWithPlanModifiers is an optional interface on Attribute which enables
// plan modification support.
type AttributeWithPlanModifiers interface {
	// Implementations should include the fwschema.Attribute interface methods
	// for proper attribute handling.
	types.Attribute

	// GetPlanModifiers should return a list of attribute-based plan modifiers.
	// This is named differently than PlanModifiers to prevent a conflict with
	// the tfsdk.Attribute field name.
	GetPlanModifiers() tfsdk.AttributePlanModifiers
}
