package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type nestingMode uint8

const (
	nestingModeSingle nestingMode = 0
	nestingModeList   nestingMode = 1
	nestingModeSet    nestingMode = 2
	nestingModeMap    nestingMode = 3
)

// NestedAttributes surfaces a group of attributes to nest beneath another
// attribute, and how that nesting should behave. Nesting can have the
// following modes:
//
// * SingleNestedAttributes are nested attributes that represent a struct or
// object; there should only be one instance of them nested beneath that
// specific attribute.
//
// * ListNestedAttributes are nested attributes that represent a list of
// structs or objects; there can be multiple instances of them beneath that
// specific attribute.
//
// * SetNestedAttributes are nested attributes that represent a set of structs
// or objects; there can be multiple instances of them beneath that specific
// attribute. Unlike ListNestedAttributes, these nested attributes must have
// unique values.
//
// * MapNestedAttributes are nested attributes that represent a string-indexed
// map of structs or objects; there can be multiple instances of them beneath
// that specific attribute. Unlike ListNestedAttributes, these nested
// attributes must be associated with a unique key. Unlike SetNestedAttributes,
// the key must be explicitly set by the user.
type NestedAttributes interface {
	getNestingMode() nestingMode
	getAttributes() map[string]Attribute
	tftypes.AttributePathStepper
}

type nestedAttributes map[string]Attribute

func (n nestedAttributes) getAttributes() map[string]Attribute {
	return map[string]Attribute(n)
}

// SingleNestedAttributes nests `attributes` under another attribute, only
// allowing one instance of that group of attributes to appear in the
// configuration.
func SingleNestedAttributes(attributes map[string]Attribute) NestedAttributes {
	return singleNestedAttributes{
		nestedAttributes(attributes),
	}
}

type singleNestedAttributes struct {
	nestedAttributes
}

func (s singleNestedAttributes) getNestingMode() nestingMode {
	return nestingModeSingle
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// nested attributes.
func (s singleNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.ElementKeyString); !ok {
		return nil, fmt.Errorf("cannot apply step %T to SingleNestedAttributes", step)
	}

	return s.nestedAttributes, nil
}

// ListNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Minimum and maximum numbers of times the group can appear in
// the configuration can be set using `opts`.
func ListNestedAttributes(attributes map[string]Attribute, opts ListNestedAttributesOptions) NestedAttributes {
	return listNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
		min:              opts.MinItems,
		max:              opts.MaxItems,
	}
}

type listNestedAttributes struct {
	nestedAttributes

	min, max int
}

// ListNestedAttributesOptions captures additional, optional parameters for
// ListNestedAttributes.
type ListNestedAttributesOptions struct {
	MinItems int
	MaxItems int
}

func (l listNestedAttributes) getNestingMode() nestingMode {
	return nestingModeList
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// nested attributes.
func (l listNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.ElementKeyInt); !ok {
		return nil, fmt.Errorf("cannot apply step %T to ListNestedAttributes", step)
	}

	return l.nestedAttributes, nil
}

// SetNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration, while requiring each group of values be unique. Minimum and
// maximum numbers of times the group can appear in the configuration can be
// set using `opts`.
func SetNestedAttributes(attributes map[string]Attribute, opts SetNestedAttributesOptions) NestedAttributes {
	return setNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
		min:              opts.MinItems,
		max:              opts.MaxItems,
	}
}

type setNestedAttributes struct {
	nestedAttributes

	min, max int
}

// SetNestedAttributesOptions captures additional, optional parameters for
// SetNestedAttributes.
type SetNestedAttributesOptions struct {
	MinItems int
	MaxItems int
}

func (s setNestedAttributes) getNestingMode() nestingMode {
	return nestingModeSet
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// nested attributes.
func (s setNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.ElementKeyInt); !ok {
		return nil, fmt.Errorf("cannot apply step %T to SetNestedAttributes", step)
	}

	return s.nestedAttributes, nil
}

// MapNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Each group will need to be associated with a unique string by
// the user. Minimum and maximum numbers of times the group can appear in the
// configuration can be set using `opts`.
func MapNestedAttributes(attributes map[string]Attribute, opts MapNestedAttributesOptions) NestedAttributes {
	return mapNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
		min:              opts.MinItems,
		max:              opts.MaxItems,
	}
}

type mapNestedAttributes struct {
	nestedAttributes

	min, max int
}

// MapNestedAttributesOptions captures additional, optional parameters for
// MapNestedAttributes.
type MapNestedAttributesOptions struct {
	MinItems int
	MaxItems int
}

func (m mapNestedAttributes) getNestingMode() nestingMode {
	return nestingModeMap
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// nested attributes.
func (m mapNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.ElementKeyString); !ok {
		return nil, fmt.Errorf("cannot apply step %T to MapNestedAttributes", step)
	}

	return m.nestedAttributes, nil
}
