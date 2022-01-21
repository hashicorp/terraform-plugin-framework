package tfsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// NestingMode is an enum type of the ways nested attributes can be nested in
// an attribute. They can be a list, a set, a map (with string
// keys), or they can be nested directly, like an object.
type NestingMode uint8

const (
	// NestingModeUnknown is an invalid nesting mode, used to catch when a
	// nesting mode is expected and not set.
	NestingModeUnknown NestingMode = 0

	// NestingModeSingle is for attributes that represent a struct or
	// object, a single instance of those attributes directly nested under
	// another attribute.
	NestingModeSingle NestingMode = 1

	// NestingModeList is for attributes that represent a list of objects,
	// with multiple instances of those attributes nested inside a list
	// under another attribute.
	NestingModeList NestingMode = 2

	// NestingModeSet is for attributes that represent a set of objects,
	// with multiple, unique instances of those attributes nested inside a
	// set under another attribute.
	NestingModeSet NestingMode = 3

	// NestingModeMap is for attributes that represent a map of objects,
	// with multiple instances of those attributes, each associated with a
	// unique string key, nested inside a map under another attribute.
	NestingModeMap NestingMode = 4
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
	tftypes.AttributePathStepper
	AttributeType() attr.Type
	GetNestingMode() NestingMode
	GetAttributes() map[string]Attribute
	Equal(NestedAttributes) bool
	unimplementable()
}

type nestedAttributes map[string]Attribute

func (n nestedAttributes) GetAttributes() map[string]Attribute {
	return map[string]Attribute(n)
}

func (n nestedAttributes) unimplementable() {}

func (n nestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	a, ok := step.(tftypes.AttributeName)
	if !ok {
		return nil, fmt.Errorf("can't apply %T to Attributes", step)
	}
	res, ok := n[string(a)]
	if !ok {
		return nil, fmt.Errorf("no attribute %q on Attributes", a)
	}
	return res, nil
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (n nestedAttributes) AttributeType() attr.Type {
	attrTypes := map[string]attr.Type{}
	for name, attr := range n.GetAttributes() {
		if attr.Type != nil {
			attrTypes[name] = attr.Type
		}
		if attr.Attributes != nil {
			attrTypes[name] = attr.Attributes.AttributeType()
		}
	}
	return types.ObjectType{
		AttrTypes: attrTypes,
	}
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

func (s singleNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeSingle
}

func (s singleNestedAttributes) Equal(o NestedAttributes) bool {
	other, ok := o.(singleNestedAttributes)
	if !ok {
		return false
	}
	if len(other.nestedAttributes) != len(s.nestedAttributes) {
		return false
	}
	for k, v := range s.nestedAttributes {
		otherV, ok := other.nestedAttributes[k]
		if !ok {
			return false
		}
		if !v.Equal(otherV) {
			return false
		}
	}
	return true
}

// ListNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Minimum and maximum numbers of times the group can appear in
// the configuration can be set using `opts`.
func ListNestedAttributes(attributes map[string]Attribute, opts ListNestedAttributesOptions) NestedAttributes {
	return listNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
	}
}

type listNestedAttributes struct {
	nestedAttributes
}

// ListNestedAttributesOptions captures additional, optional parameters for
// ListNestedAttributes.
type ListNestedAttributesOptions struct{}

func (l listNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeList
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (l listNestedAttributes) AttributeType() attr.Type {
	return types.ListType{
		ElemType: l.nestedAttributes.AttributeType(),
	}
}

func (l listNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyInt)
	if !ok {
		return nil, fmt.Errorf("can't apply %T to ListNestedAttributes", step)
	}
	return l.nestedAttributes, nil
}

func (l listNestedAttributes) Equal(o NestedAttributes) bool {
	other, ok := o.(listNestedAttributes)
	if !ok {
		return false
	}
	if len(other.nestedAttributes) != len(l.nestedAttributes) {
		return false
	}
	for k, v := range l.nestedAttributes {
		otherV, ok := other.nestedAttributes[k]
		if !ok {
			return false
		}
		if !v.Equal(otherV) {
			return false
		}
	}
	return true
}

// SetNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration, while requiring each group of values be unique. Minimum and
// maximum numbers of times the group can appear in the configuration can be
// set using `opts`.
func SetNestedAttributes(attributes map[string]Attribute, opts SetNestedAttributesOptions) NestedAttributes {
	return setNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
	}
}

type setNestedAttributes struct {
	nestedAttributes
}

// SetNestedAttributesOptions captures additional, optional parameters for
// SetNestedAttributes.
type SetNestedAttributesOptions struct{}

func (s setNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeSet
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (s setNestedAttributes) AttributeType() attr.Type {
	return types.SetType{
		ElemType: s.nestedAttributes.AttributeType(),
	}
}

func (s setNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyValue)
	if !ok {
		return nil, fmt.Errorf("can't use %T on sets", step)
	}
	return s.nestedAttributes, nil
}

func (s setNestedAttributes) Equal(o NestedAttributes) bool {
	other, ok := o.(setNestedAttributes)
	if !ok {
		return false
	}
	if len(other.nestedAttributes) != len(s.nestedAttributes) {
		return false
	}
	for k, v := range s.nestedAttributes {
		otherV, ok := other.nestedAttributes[k]
		if !ok {
			return false
		}
		if !v.Equal(otherV) {
			return false
		}
	}
	return true
}

// MapNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Each group will need to be associated with a unique string by
// the user. Minimum and maximum numbers of times the group can appear in the
// configuration can be set using `opts`.
func MapNestedAttributes(attributes map[string]Attribute, opts MapNestedAttributesOptions) NestedAttributes {
	return mapNestedAttributes{
		nestedAttributes: nestedAttributes(attributes),
	}
}

type mapNestedAttributes struct {
	nestedAttributes
}

// MapNestedAttributesOptions captures additional, optional parameters for
// MapNestedAttributes.
type MapNestedAttributesOptions struct{}

func (m mapNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeMap
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (m mapNestedAttributes) AttributeType() attr.Type {
	return types.MapType{
		ElemType: m.nestedAttributes.AttributeType(),
	}
}

func (m mapNestedAttributes) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyString)
	if !ok {
		return nil, fmt.Errorf("can't use %T on maps", step)
	}
	return m.nestedAttributes, nil
}

func (m mapNestedAttributes) Equal(o NestedAttributes) bool {
	other, ok := o.(mapNestedAttributes)
	if !ok {
		return false
	}
	if len(other.nestedAttributes) != len(m.nestedAttributes) {
		return false
	}
	for k, v := range m.nestedAttributes {
		otherV, ok := other.nestedAttributes[k]
		if !ok {
			return false
		}
		if !v.Equal(otherV) {
			return false
		}
	}
	return true
}
