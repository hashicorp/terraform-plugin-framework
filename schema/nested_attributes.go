package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestingMode uint8

const (
	NestingModeUnknown NestingMode = 0
	NestingModeSingle  NestingMode = 1
	NestingModeList    NestingMode = 2
	NestingModeSet     NestingMode = 3
	NestingModeMap     NestingMode = 4
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
	AttributeType() attr.Type
	GetNestingMode() NestingMode
	GetAttributes() map[string]Attribute
	GetMinItems() int64
	GetMaxItems() int64
	unimplementable()
}

type nestedAttributes map[string]Attribute

func (n nestedAttributes) GetAttributes() map[string]Attribute {
	return map[string]Attribute(n)
}

func (n nestedAttributes) unimplementable() {}

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

func (s singleNestedAttributes) GetMinItems() int64 {
	return 0
}

func (s singleNestedAttributes) GetMaxItems() int64 {
	return 0
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (s singleNestedAttributes) AttributeType() attr.Type {
	attrTypes := map[string]attr.Type{}
	for name, attr := range s.GetAttributes() {
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

func (l listNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeList
}

func (l listNestedAttributes) GetMinItems() int64 {
	return int64(l.min)
}

func (l listNestedAttributes) GetMaxItems() int64 {
	return int64(l.max)
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (l listNestedAttributes) AttributeType() attr.Type {
	attrTypes := map[string]attr.Type{}
	for name, attr := range l.GetAttributes() {
		if attr.Type != nil {
			attrTypes[name] = attr.Type
		}
		if attr.Attributes != nil {
			attrTypes[name] = attr.Attributes.AttributeType()
		}
	}
	return types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: attrTypes,
		},
	}
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

func (s setNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeSet
}

func (s setNestedAttributes) GetMinItems() int64 {
	return int64(s.min)
}

func (s setNestedAttributes) GetMaxItems() int64 {
	return int64(s.max)
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (s setNestedAttributes) AttributeType() attr.Type {
	// TODO fill in implementation when types.SetType is available
	return nil
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

func (m mapNestedAttributes) GetNestingMode() NestingMode {
	return NestingModeMap
}

func (m mapNestedAttributes) GetMinItems() int64 {
	return int64(m.min)
}

func (m mapNestedAttributes) GetMaxItems() int64 {
	return int64(m.max)
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (m mapNestedAttributes) AttributeType() attr.Type {
	// TODO fill in implementation when types.MapType is available
	return nil
}
