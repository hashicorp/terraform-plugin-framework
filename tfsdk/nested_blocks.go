package tfsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// NestedBlocks surfaces a group of attributes to nest beneath another
// attribute as a block, and how that nesting should behave. Block nesting can
// have the following modes:
//
// * ListNestedBlocks are nested attributes as a block that represent a list of
// structs or objects; there can be multiple instances of them beneath that
// specific attribute.
//
// * SetNestedBlocks are nested attributes as a block that represent a set of
//  structs or objects; there can be multiple instances of them beneath that
// specific attribute. Unlike ListNestedBlocks, these nested attributes must have
// unique values.
type NestedBlocks interface {
	tftypes.AttributePathStepper
	AttributeType() attr.Type
	GetNestingMode() NestingMode
	GetAttributes() map[string]Attribute
	GetMinItems() int64
	GetMaxItems() int64
	Equal(NestedBlocks) bool
	unimplementable()
}

// ListNestedBlocks nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Minimum and maximum numbers of times the group can appear in
// the configuration can be set using `opts`.
func ListNestedBlocks(attributes map[string]Attribute, opts ListNestedBlocksOptions) NestedBlocks {
	return listNestedBlocks{
		nestedAttributes: nestedAttributes(attributes),
		min:              opts.MinItems,
		max:              opts.MaxItems,
	}
}

type listNestedBlocks struct {
	nestedAttributes

	min, max int
}

// ListNestedBlocksOptions captures additional, optional parameters for
// ListNestedBlocks.
type ListNestedBlocksOptions struct {
	MinItems int
	MaxItems int
}

func (l listNestedBlocks) GetNestingMode() NestingMode {
	return NestingModeList
}

func (l listNestedBlocks) GetMinItems() int64 {
	return int64(l.min)
}

func (l listNestedBlocks) GetMaxItems() int64 {
	return int64(l.max)
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (l listNestedBlocks) AttributeType() attr.Type {
	return types.ListType{
		ElemType: l.nestedAttributes.AttributeType(),
	}
}

func (l listNestedBlocks) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyInt)
	if !ok {
		return nil, fmt.Errorf("can't apply %T to ListNestedBlocks", step)
	}
	return l.nestedAttributes, nil
}

func (l listNestedBlocks) Equal(o NestedBlocks) bool {
	other, ok := o.(listNestedBlocks)
	if !ok {
		return false
	}
	if l.min != other.min {
		return false
	}
	if l.max != other.max {
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

// SetNestedBlocks nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration, while requiring each group of values be unique. Minimum and
// maximum numbers of times the group can appear in the configuration can be
// set using `opts`.
func SetNestedBlocks(attributes map[string]Attribute, opts SetNestedBlocksOptions) NestedBlocks {
	return setNestedBlocks{
		nestedAttributes: nestedAttributes(attributes),
		min:              opts.MinItems,
		max:              opts.MaxItems,
	}
}

type setNestedBlocks struct {
	nestedAttributes

	min, max int
}

// SetNestedBlocksOptions captures additional, optional parameters for
// SetNestedBlocks.
type SetNestedBlocksOptions struct {
	MinItems int
	MaxItems int
}

func (s setNestedBlocks) GetNestingMode() NestingMode {
	return NestingModeSet
}

func (s setNestedBlocks) GetMinItems() int64 {
	return int64(s.min)
}

func (s setNestedBlocks) GetMaxItems() int64 {
	return int64(s.max)
}

// AttributeType returns an attr.Type corresponding to the nested attributes.
func (s setNestedBlocks) AttributeType() attr.Type {
	return types.SetType{
		ElemType: s.nestedAttributes.AttributeType(),
	}
}

func (s setNestedBlocks) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyValue)
	if !ok {
		return nil, fmt.Errorf("can't use %T on sets", step)
	}
	return s.nestedAttributes, nil
}

func (s setNestedBlocks) Equal(o NestedBlocks) bool {
	other, ok := o.(setNestedBlocks)
	if !ok {
		return false
	}
	if s.min != other.min {
		return false
	}
	if s.max != other.max {
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
