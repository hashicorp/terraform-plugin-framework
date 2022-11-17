package fwxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// BlockWithValidators is an optional interface on Block which enables
// validation support.
type BlockWithValidators interface {
	// Implementations should include the fwschema.Block interface methods
	// for proper block handling.
	fwschema.Block

	// GetValidators should return a list of attribute-based validators. This
	// is named differently than Validators to prevent a conflict with the
	// tfsdk.Block field name.
	GetValidators() []tfsdk.AttributeValidator
}

// BlockWithListValidators is an optional interface on Block which
// enables List validation support.
type BlockWithListValidators interface {
	fwschema.Block

	// ListValidators should return a list of List validators.
	ListValidators() []validator.List
}

// BlockWithObjectValidators is an optional interface on Block which
// enables Object validation support.
type BlockWithObjectValidators interface {
	fwschema.Block

	// ObjectValidators should return a list of Object validators.
	ObjectValidators() []validator.Object
}

// BlockWithSetValidators is an optional interface on Block which
// enables Set validation support.
type BlockWithSetValidators interface {
	fwschema.Block

	// SetValidators should return a list of Set validators.
	SetValidators() []validator.Set
}
