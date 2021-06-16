package path

import "github.com/hashicorp/terraform-plugin-framework/attr"

// AttributePath is used to identify part of a schema.Schema, attr.Type, or
// attr.Value. It is a series of steps, describing a traversal of a Terraform
// type. Each Terraform type has its own limits on what steps can be used to
// traverse it.
type AttributePath struct {
	steps []step
}

// New returns an AttributePath that is ready to be used.
func New() AttributePath {
	return AttributePath{}
}

func copySteps(in []step) []step {
	out := make([]step, len(in))
	copy(out, in)
	return out
}

// Attribute defensively copies the existing steps on the AttributePath, and
// returns a new AttributePath containing those steps, with the specified
// AttributeName step added to the end.
func (a AttributePath) Attribute(name string) AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), AttributeName(name)),
	}
}

// WildAttribute defensively copies the existing steps on the AttributePath,
// and returns a new AttributePath containing those steps, with a
// WildAttributeName step added to the end.
func (a AttributePath) WildAttribute() AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), WildAttributeName{}),
	}
}

// StringElement defensively copies the existing steps on the AttributePath,
// and returns a new AttributePath containing those steps, with the specified
// ElementKeyString step added to the end.
func (a AttributePath) StringElement(name string) AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), ElementKeyString(name)),
	}
}

// WildStringElement defensively copies the existing steps on the
// AttributePath, and returns a new AttributePath containing those steps, with
// a WildElementKeyString step added to the end.
func (a AttributePath) WildStringElement() AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), WildElementKeyString{}),
	}
}

// IntElement defensively copies the existing steps on the AttributePath, and
// returns a new AttributePath containing those steps, with the specified
// ElementKeyInt step added to the end.
func (a AttributePath) IntElement(pos int) AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), ElementKeyInt(pos)),
	}
}

// WildIntElement defensively copies the existing steps on the AttributePath,
// and returns a new AttributePath containing those steps, with a
// WildElementKeyInt step added to the end.
func (a AttributePath) WildIntElement() AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), WildElementKeyInt{}),
	}
}

// ValueElement defensively copies the existing steps on the AttributePath, and
// returns a new AttributePath containing those steps, with the specified
// ElementKeyValue step added to the end.
func (a AttributePath) ValueElement(val attr.Value) AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), ElementKeyValue{Value: val}),
	}
}

// WildValueElement defensively copies the existing steps on the AttributePath,
// and returns a new AttributePath containing those steps, with a
// WildElementKeyValue step added to the end.
func (a AttributePath) WildValueElement() AttributePath {
	return AttributePath{
		steps: append(copySteps(a.steps), WildElementKeyValue{}),
	}
}

type step interface {
	unimplementable()
}

// AttributeName is a step that selects a single attribute, by its name. It is
// used on attr.TypeWithAttributeTypes, schema.NestedAttributes, and
// schema.Schema types.
type AttributeName string

func (a AttributeName) unimplementable() {}

// ElementKeyString is a step that selects a single element using a string
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Map from their TerraformType method. It is also used on
// schema.MapNestedAttribute nested attributes.
type ElementKeyString string

func (e ElementKeyString) unimplementable() {}

// ElementKeyInt is a step that selects a single element using an integer
// index. It is used on attr.TypeWithElementTypes types and
// attr.TypeWithElementType types that return a tftypes.List from their
// TerraformType method. It is also used on schema.ListNestedAttribute nested
// attributes.
type ElementKeyInt int

func (e ElementKeyInt) unimplementable() {}

// ElementKeyValue is a step that selects a single element using an attr.Value
// index. It is used on attr.TypeWithElementType types that return a
// tftypes.Set from their TerraformType method. It is also used on
// schema.SetNestedAttribute nested attributes.
type ElementKeyValue struct {
	Value attr.Value
}

func (e ElementKeyValue) unimplementable() {}

// WildAttributeName is used as AttributeName is, except instead of matching a
// specific attribute name, it acts as a wildcard, matching any attribute name.
// This can't be used in diagnostics, it is only useful for validation.
type WildAttributeName struct{}

func (w WildAttributeName) unimplementable() {}

// WildElementKeyString is used as ElementKeyString is, except instead of
// matching a specific element key, it acts as a wildcard, matching any element
// key. This can't be used in diagnostics, it is only useful for validation.
type WildElementKeyString struct{}

func (w WildElementKeyString) unimplementable() {}

// WildElementKeyInt is used as ElementKeyInt is, except instead of matching a
// specific element key, it acts as a wildcard, matching any element key. This
// can't be used in diagnostics, it is only useful for validation.
type WildElementKeyInt struct{}

func (w WildElementKeyInt) unimplementable() {}

// WildElementKeyValue is used as ElementKeyValue is, except instead of
// matching a specific element key, it acts as a wildcard, matching any element
// key. This can't be used in diagnostics, it is only useful for validation.
type WildElementKeyValue struct{}

func (w WildElementKeyValue) unimplementable() {}
