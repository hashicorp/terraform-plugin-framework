package fwxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// BlockWithPlanModifiers is an optional interface on Block which enables
// plan modification support.
type BlockWithPlanModifiers interface {
	// Implementations should include the fwschema.Block interface methods
	// for proper block handling.
	fwschema.Block

	// GetPlanModifiers should return a list of attribute-based plan modifiers.
	// This is named differently than PlanModifiers to prevent a conflict with
	// the tfsdk.Block field name.
	GetPlanModifiers() tfsdk.AttributePlanModifiers
}

// BlockWithListPlanModifiers is an optional interface on Block which
// enables List plan modifier support.
type BlockWithListPlanModifiers interface {
	fwschema.Block

	// ListPlanModifiers should return a list of List plan modifiers.
	ListPlanModifiers() []planmodifier.List
}

// BlockWithObjectPlanModifiers is an optional interface on Block which
// enables Object plan modifier support.
type BlockWithObjectPlanModifiers interface {
	fwschema.Block

	// ObjectPlanModifiers should return a list of Object plan modifiers.
	ObjectPlanModifiers() []planmodifier.Object
}

// BlockWithSetPlanModifiers is an optional interface on Block which
// enables Set plan modifier support.
type BlockWithSetPlanModifiers interface {
	fwschema.Block

	// SetPlanModifiers should return a list of Set plan modifiers.
	SetPlanModifiers() []planmodifier.Set
}
