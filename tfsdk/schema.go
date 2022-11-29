package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	// ErrPathInsideAtomicAttribute is used with AttributeAtPath is called
	// on a path that doesn't have a schema associated with it, because
	// it's an element, attribute, or block of a complex type, not a nested
	// attribute.
	//
	// Deprecated: This error value was intended for internal usage and will
	// be removed in a future version. If you have a use case for this,
	// please create a GitHub issue.
	ErrPathInsideAtomicAttribute = fwschema.ErrPathInsideAtomicAttribute

	// ErrPathIsBlock is used with AttributeAtPath is called on a path is a
	// block, not an attribute. Use blockAtPath on the path instead.
	//
	// Deprecated: This error value was intended for internal usage and will
	// be removed in a future version. If you have a use case for this,
	// please create a GitHub issue.
	ErrPathIsBlock = fwschema.ErrPathIsBlock
)

// Schema must satify the fwschema.Schema interface.
var _ fwschema.Schema = Schema{}

// Schema is used to define the shape of practitioner-provider information,
// like resources, data sources, and providers. Think of it as a type
// definition, but for Terraform.
//
// Deprecated: Use datasource/schema.Schema, provider/schema.Schema, or
// resource/schema.Schema instead. This can be switched by using the
// datasource.DataSource, provider.Provider, or resource.Resource interface
// Schema method.
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
	// Attributes are preferred over Blocks. Blocks should typically be used
	// for configuration compatibility with previously existing schemas from
	// an older Terraform Plugin SDK. Efforts should be made to convert from
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
	return fwschema.SchemaApplyTerraform5AttributePathStep(s, step)
}

// TypeAtPath returns the framework type at the given schema path.
func (s Schema) TypeAtPath(ctx context.Context, p path.Path) (attr.Type, diag.Diagnostics) {
	return fwschema.SchemaTypeAtPath(ctx, s, p)
}

// TypeAtTerraformPath returns the framework type at the given tftypes path.
func (s Schema) TypeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (attr.Type, error) {
	return fwschema.SchemaTypeAtTerraformPath(ctx, s, p)
}

// GetAttributes satisfies the fwschema.Schema interface.
func (s Schema) GetAttributes() map[string]fwschema.Attribute {
	return schemaAttributes(s.Attributes)
}

// GetBlocks satisfies the fwschema.Schema interface.
func (s Schema) GetBlocks() map[string]fwschema.Block {
	return schemaBlocks(s.Blocks)
}

// GetDeprecationMessage satisfies the fwschema.Schema interface.
func (s Schema) GetDeprecationMessage() string {
	return s.DeprecationMessage
}

// GetDescription satisfies the fwschema.Schema interface.
func (s Schema) GetDescription() string {
	return s.Description
}

// GetMarkdownDescription satisfies the fwschema.Schema interface.
func (s Schema) GetMarkdownDescription() string {
	return s.MarkdownDescription
}

// GetVersion satisfies the fwschema.Schema interface.
func (s Schema) GetVersion() int64 {
	return s.Version
}

// Type returns the framework type of the schema.
func (s Schema) Type() attr.Type {
	return fwschema.SchemaType(s)
}

// AttributeAtPath returns the Attribute at the passed path. If the path points
// to an element or attribute of a complex type, rather than to an Attribute,
// it will return an ErrPathInsideAtomicAttribute error.
func (s Schema) AttributeAtPath(ctx context.Context, p path.Path) (fwschema.Attribute, diag.Diagnostics) {
	return fwschema.SchemaAttributeAtPath(ctx, s, p)
}

// AttributeAtPath returns the Attribute at the passed path. If the path points
// to an element or attribute of a complex type, rather than to an Attribute,
// it will return an ErrPathInsideAtomicAttribute error.
func (s Schema) AttributeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (fwschema.Attribute, error) {
	return fwschema.SchemaAttributeAtTerraformPath(ctx, s, p)
}

// schemaAttributes is a tfsdk to fwschema type conversion function.
func schemaAttributes(attributes map[string]Attribute) map[string]fwschema.Attribute {
	result := make(map[string]fwschema.Attribute, len(attributes))

	for name, attribute := range attributes {
		result[name] = attribute
	}

	return result
}

// schemaBlocks is a tfsdk to fwschema type conversion function.
func schemaBlocks(blocks map[string]Block) map[string]fwschema.Block {
	result := make(map[string]fwschema.Block, len(blocks))

	for name, block := range blocks {
		result[name] = block
	}

	return result
}
