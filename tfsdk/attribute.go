package tfsdk

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Attribute defines the constraints and behaviors of a single value field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
	// Type indicates what kind of attribute this is. You'll most likely
	// want to use one of the types in the types package.
	//
	// If Type is set, Attributes cannot be.
	Type attr.Type

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// If Attributes is set, Type cannot be.
	Attributes NestedAttributes

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

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose not to enter
	// a value for this attribute or not. Optional and Required cannot both
	// be true.
	//
	// When defining an attribute that has Optional set to true,
	// and uses PlanModifiers to set a "default value" when none is provided,
	// Computed must also be set to true. This is necessary because default
	// values are, in effect, set by the provider (i.e. computed).
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this Attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	//
	// When defining an Optional Attribute that has a "default value"
	// plan modifier, Computed must also be set to true. Otherwise,
	// Terraform will return an error like:
	//
	//      planned value ... for a non-computed attribute
	//
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
	Sensitive bool

	// DeprecationMessage defines a message to display to practitioners
	// using this attribute, warning them that it is deprecated and
	// instructing them on what upgrade steps to take.
	DeprecationMessage string

	// Validators defines validation functionality for the attribute.
	Validators []AttributeValidator

	// PlanModifiers defines a sequence of modifiers for this attribute at
	// plan time. Attribute-level plan modifications occur before any
	// resource-level plan modifications.
	//
	// Any errors will prevent further execution of this sequence
	// of modifiers and modifiers associated with any nested Attribute, but
	// will not prevent execution of PlanModifiers on any other Attribute or
	// Block in the Schema.
	//
	// Plan modification only applies to resources, not data sources or
	// providers. Setting PlanModifiers on a data source or provider attribute
	// will have no effect.
	//
	// When providing PlanModifiers, it's necessary to set Computed to true.
	PlanModifiers AttributePlanModifiers
}

// ApplyTerraform5AttributePathStep transparently calls
// ApplyTerraform5AttributePathStep on a.Type or a.Attributes, whichever is
// non-nil. It allows Attributes to be walked using tftypes.Walk and
// tftypes.Transform.
func (a Attribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if a.Type != nil {
		return a.Type.ApplyTerraform5AttributePathStep(step)
	}
	if a.Attributes != nil {
		return a.Attributes.ApplyTerraform5AttributePathStep(step)
	}
	return nil, errors.New("Attribute has no type or nested attributes")
}

// Equal returns true if `a` and `o` should be considered Equal.
func (a Attribute) Equal(o Attribute) bool {
	if a.Type == nil && o.Type != nil {
		return false
	} else if a.Type != nil && o.Type == nil {
		return false
	} else if a.Type != nil && o.Type != nil && !a.Type.Equal(o.Type) {
		return false
	}
	if a.Attributes == nil && o.Attributes != nil {
		return false
	} else if a.Attributes != nil && o.Attributes == nil {
		return false
	} else if a.Attributes != nil && o.Attributes != nil && !a.Attributes.Equal(o.Attributes) {
		return false
	}
	if a.Description != o.Description {
		return false
	}
	if a.MarkdownDescription != o.MarkdownDescription {
		return false
	}
	if a.Required != o.Required {
		return false
	}
	if a.Optional != o.Optional {
		return false
	}
	if a.Computed != o.Computed {
		return false
	}
	if a.Sensitive != o.Sensitive {
		return false
	}
	if a.DeprecationMessage != o.DeprecationMessage {
		return false
	}
	return true
}

// attributeType returns an attr.Type corresponding to the attribute.
func (a Attribute) attributeType() attr.Type {
	if a.Attributes != nil {
		return a.Attributes.AttributeType()
	}

	return a.Type
}

// terraformType returns an tftypes.Type corresponding to the attribute.
func (a Attribute) terraformType(ctx context.Context) tftypes.Type {
	if a.Attributes != nil {
		return a.Attributes.AttributeType().TerraformType(ctx)
	}

	return a.Type.TerraformType(ctx)
}
