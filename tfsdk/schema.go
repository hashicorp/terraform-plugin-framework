package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	// ErrPathInsideAtomicAttribute is used with AttributeAtPath is called
	// on a path that doesn't have a schema associated with it, because
	// it's an element, attribute, or block of a complex type, not a nested
	// attribute.
	ErrPathInsideAtomicAttribute = errors.New("path leads to element, attribute, or block of a schema.Attribute that has no schema associated with it")

	// ErrPathIsBlock is used with AttributeAtPath is called on a path is a
	// block, not an attribute. Use blockAtPath on the path instead.
	ErrPathIsBlock = errors.New("path leads to block, not an attribute")
)

// Schema is used to define the shape of practitioner-provider information,
// like resources, data sources, and providers. Think of it as a type
// definition, but for Terraform.
type Schema struct {
	// Attributes are value fields inside the resource, provider, or data
	// source that the schema is defining. The map key should be the name
	// of the attribute, and the body defines how it behaves. Names must
	// only contain lowercase letters, numbers, and underscores. Names must
	// not collide with any Blocks names.
	//
	// In practitioner configurations, an equals sign (=) is required to set
	// the value. See also:
	//   https://www.terraform.io/docs/language/syntax/configuration.html
	//
	// Attributes are strongly preferred over Blocks.
	Attributes map[string]Attribute

	// Blocks are structural fields inside the resource, provider, or data
	// source that the schema is defining. The map key should be the name
	// of the block, and the body defines how it behaves. Names must
	// only contain lowercase letters, numbers, and underscores. Names must
	// not collide with any Attributes names.
	//
	// Blocks are by definition, structural, meaning they are implicitly
	// required in values.
	//
	// In practitioner configurations, an equals sign (=) cannot be used to
	// set the value. Blocks are instead repeated as necessary, or require
	// the use of dynamic block expressions. See also:
	//   https://www.terraform.io/docs/language/syntax/configuration.html
	//   https://www.terraform.io/docs/language/expressions/dynamic-blocks.html
	//
	// Deprecated: Attributes are preferred over Blocks. Blocks should only be
	// used for configuration compatibility with previously existing schemas
	// from an older Terraform Plugin SDK. Efforts should be made to convert
	// Blocks to Attributes as a breaking change for practitioners.
	Blocks map[string]Block

	// Version indicates the current version of the schema. Schemas are
	// versioned to help with automatic upgrade process. This is not
	// typically required unless there is a change in the schema, such as
	// changing an attribute type, that needs manual upgrade handling.
	// Versions should only be incremented by one each release.
	Version int64

	DeprecationMessage  string
	Description         string
	MarkdownDescription string
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// schema.
func (s Schema) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	a, ok := step.(tftypes.AttributeName)

	if !ok {
		return nil, fmt.Errorf("cannot apply AttributePathStep %T to schema", step)
	}

	attrName := string(a)

	if attr, ok := s.Attributes[attrName]; ok {
		return attr, nil
	}

	if block, ok := s.Blocks[attrName]; ok {
		return block, nil
	}

	return nil, fmt.Errorf("could not find attribute or block %q in schema", a)
}

// AttributeType returns a types.ObjectType composed from the schema types.
func (s Schema) AttributeType() attr.Type {
	attrTypes := map[string]attr.Type{}
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.attributeType()
	}
	for name, block := range s.Blocks {
		attrTypes[name] = block.attributeType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}

// AttributeTypeAtPath returns the attr.Type of the attribute at the given path.
func (s Schema) AttributeTypeAtPath(path *tftypes.AttributePath) (attr.Type, error) {
	rawType, remaining, err := tftypes.WalkAttributePath(s, path)
	if err != nil {
		return nil, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}

	switch typ := rawType.(type) {
	case attr.Type:
		return typ, nil
	case nestedAttributes:
		return typ.AttributeType(), nil
	case nestedBlock:
		return typ.Block.attributeType(), nil
	case Attribute:
		return typ.attributeType(), nil
	case Block:
		return typ.attributeType(), nil
	case Schema:
		return typ.AttributeType(), nil
	default:
		return nil, fmt.Errorf("got unexpected type %T", rawType)
	}
}

// TerraformType returns a tftypes.Type that can represent the schema.
func (s Schema) TerraformType(ctx context.Context) tftypes.Type {
	attrTypes := map[string]tftypes.Type{}
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.terraformType(ctx)
	}
	for name, block := range s.Blocks {
		attrTypes[name] = block.terraformType(ctx)
	}
	return tftypes.Object{AttributeTypes: attrTypes}
}

// AttributeAtPath returns the Attribute at the passed path. If the path points
// to an element or attribute of a complex type, rather than to an Attribute,
// it will return an ErrPathInsideAtomicAttribute error.
func (s Schema) AttributeAtPath(path *tftypes.AttributePath) (Attribute, error) {
	res, remaining, err := tftypes.WalkAttributePath(s, path)
	if err != nil {
		return Attribute{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}

	switch r := res.(type) {
	case attr.Type:
		return Attribute{}, ErrPathInsideAtomicAttribute
	case nestedAttributes:
		return Attribute{}, ErrPathInsideAtomicAttribute
	case nestedBlock:
		return Attribute{}, ErrPathInsideAtomicAttribute
	case Attribute:
		return r, nil
	case Block:
		return Attribute{}, ErrPathIsBlock
	default:
		return Attribute{}, fmt.Errorf("got unexpected type %T", res)
	}
}
