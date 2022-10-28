package fwxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeWithValidators is an optional interface on Attribute which enables
// validation support.
type AttributeWithValidators interface {
	// Implementations should include the fwschema.Attribute interface methods
	// for proper attribute handling.
	types.Attribute

	// GetValidators should return a list of attribute-based validators. This
	// is named differently than PlanModifiers to prevent a conflict with the
	// tfsdk.Attribute field name.
	GetValidators() []tfsdk.AttributeValidator
}
