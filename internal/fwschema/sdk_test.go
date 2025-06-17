package fwschema_test

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFromSDK(t *testing.T) { //nolint:paralleltest
	sdkResourceSchema := sdkschema.Resource{
		Schema: map[string]*sdkschema.Schema{
			"location": {
				Type:     sdkschema.TypeString,
				Required: true,
			},
			"cpu": {
				Type:     sdkschema.TypeInt,
				Optional: true,
			},
			"disk": {
				Type: sdkschema.TypeList,
				Elem: &sdkschema.Resource{
					Schema: map[string]*sdkschema.Schema{
						"capacity": {
							Type:     sdkschema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
	schema := NewSDKSchema(sdkResourceSchema)
	attributes := schema.GetAttributes()

	if len(attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %v", slices.Collect(maps.Keys(attributes)))
	}

	if !attributes["location"].IsRequired() {
		t.Fatalf("expected location to be required")
	}

	if attributes["location"].IsOptional() {
		t.Fatalf("expected location not to be optional")
	}

	if attributes["location"].IsComputed() {
		t.Fatalf("expected location not to be computed")
	}

	if !attributes["cpu"].IsOptional() {
		t.Fatalf("expected cpu to be optional")
	}

	blocks := schema.GetBlocks()

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %v", slices.Collect(maps.Keys(blocks)))
	}
}

var _ fwschema.Schema = &SDKSchema{}
var _ fwschema.Attribute = &SDKAttribute{}
var _ fwschema.Block = &SDKBlock{}

func NewSDKSchema(sdkResourceSchema sdkschema.Resource) fwschema.Schema {
	return &SDKSchema{sdkResourceSchema: &sdkResourceSchema}
}

type SDKSchema struct {
	sdkResourceSchema *sdkschema.Resource
}

type SDKAttribute struct {
	sdkSchema *sdkschema.Schema
}

type SDKBlock struct {
	sdkSchema *sdkschema.Schema
}

// Return the attribute or element the AttributePathStep is referring
// to, or an error if the AttributePathStep is referring to an
// attribute or element that doesn't exist.
func (s *SDKBlock) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	return nil, nil
}

// Equal should return true if the other block is exactly equivalent.
func (s *SDKBlock) Equal(o fwschema.Block) bool {
	return false
}

// GetDeprecationMessage should return a non-empty string if an attribute
// is deprecated. This is named differently than DeprecationMessage to
// prevent a conflict with the tfsdk.Attribute field name.
func (s *SDKBlock) GetDeprecationMessage() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetDescription should return a non-empty string if an attribute
// has a plaintext description. This is named differently than Description
// to prevent a conflict with the tfsdk.Attribute field name.
func (s *SDKBlock) GetDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetMarkdownDescription should return a non-empty string if an attribute
// has a Markdown description. This is named differently than
// MarkdownDescription to prevent a conflict with the tfsdk.Attribute field
// name.
func (s *SDKBlock) GetMarkdownDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetNestedObject should return the object underneath the block.
// For single nesting mode, the NestedBlockObject can be generated from
// the Block.
func (s *SDKBlock) GetNestedObject() fwschema.NestedBlockObject {
	return nil
}

// GetNestingMode should return the nesting mode of a block. This is named
// differently than NestingMode to prevent a conflict with the tfsdk.Block
// field name.
func (s *SDKBlock) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeUnknown
}

// Type should return the framework type of a block.
func (s *SDKBlock) Type() attr.Type {
	return nil
}

// Return the attribute or element the AttributePathStep is referring
// to, or an error if the AttributePathStep is referring to an
// attribute or element that doesn't exist.
func (s *SDKAttribute) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	return nil, nil
}

// Equal should return true if the other attribute is exactly equivalent.
func (s *SDKAttribute) Equal(o fwschema.Attribute) bool {
	return false
}

// GetDeprecationMessage should return a non-empty string if an attribute
// is deprecated. This is named differently than DeprecationMessage to
// prevent a conflict with the tfsdk.Attribute field name.
func (s *SDKAttribute) GetDeprecationMessage() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetDescription should return a non-empty string if an attribute
// has a plaintext description. This is named differently than Description
// to prevent a conflict with the tfsdk.Attribute field name.
func (s *SDKAttribute) GetDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetMarkdownDescription should return a non-empty string if an attribute
// has a Markdown description. This is named differently than
// MarkdownDescription to prevent a conflict with the tfsdk.Attribute field
// name.
func (s *SDKAttribute) GetMarkdownDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetType should return the framework type of an attribute. This is named
// differently than Type to prevent a conflict with the tfsdk.Attribute
// field name.
func (s *SDKAttribute) GetType() attr.Type {
	return nil
}

// IsComputed should return true if the attribute configuration value is
// computed. This is named differently than Computed to prevent a conflict
// with the tfsdk.Attribute field name.
func (s *SDKAttribute) IsComputed() bool {
	return false
}

// IsOptional should return true if the attribute configuration value is
// optional. This is named differently than Optional to prevent a conflict
// with the tfsdk.Attribute field name.
func (s *SDKAttribute) IsOptional() bool {
	return s.sdkSchema.Optional
}

// IsRequired should return true if the attribute configuration value is
// required. This is named differently than Required to prevent a conflict
// with the tfsdk.Attribute field name.
func (s *SDKAttribute) IsRequired() bool {
	return s.sdkSchema.Required
}

// IsSensitive should return true if the attribute configuration value is
// sensitive. This is named differently than Sensitive to prevent a
// conflict with the tfsdk.Attribute field name.
func (s *SDKAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly should return true if the attribute configuration value is
// write-only. This is named differently than WriteOnly to prevent a
// conflict with the tfsdk.Attribute field name.
//
// Write-only attributes are a managed-resource schema concept only.
func (s *SDKAttribute) IsWriteOnly() bool {
	return false
}

// IsOptionalForImport should return true if the identity attribute is optional to be set by
// the practitioner when importing by identity. This is named differently than OptionalForImport
// to prevent a conflict with the relevant field name.
func (s *SDKAttribute) IsOptionalForImport() bool {
	return false
}

// IsRequiredForImport should return true if the identity attribute must be set by
// the practitioner when importing by identity. This is named differently than RequiredForImport
// to prevent a conflict with the relevant field name.
func (s *SDKAttribute) IsRequiredForImport() bool {
	return false
}

// Return the attribute or element the AttributePathStep is referring
// to, or an error if the AttributePathStep is referring to an
// attribute or element that doesn't exist.
func (s *SDKSchema) ApplyTerraform5AttributePathStep(_ tftypes.AttributePathStep) (interface{}, error) {
	return nil, nil
}

// AttributeAtPath should return the Attribute at the given path or return
// an error.
func (s *SDKSchema) AttributeAtPath(_ context.Context, _ path.Path) (fwschema.Attribute, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// AttributeAtTerraformPath should return the Attribute at the given
// Terraform path or return an error.
func (s *SDKSchema) AttributeAtTerraformPath(_ context.Context, _ *tftypes.AttributePath) (fwschema.Attribute, error) {
	return nil, nil
}

// GetAttributes should return the attributes of a schema. This is named
// differently than Attributes to prevent a conflict with the tfsdk.Schema
// field name.
func (s *SDKSchema) GetAttributes() map[string]fwschema.Attribute {
	attributes := make(map[string]fwschema.Attribute)

	schemaMap := s.sdkResourceSchema.Schema
	for name, sdkAttr := range schemaMap {
		switch sdkAttr.Type {
		case sdkschema.TypeInt, sdkschema.TypeString:
			attributes[name] = &SDKAttribute{sdkSchema: sdkAttr}
		}
	}

	return attributes
}

// GetBlocks should return the blocks of a schema. This is named
// differently than Blocks to prevent a conflict with the tfsdk.Schema
// field name.
func (s *SDKSchema) GetBlocks() map[string]fwschema.Block {
	blocks := make(map[string]fwschema.Block)

	schemaMap := s.sdkResourceSchema.Schema
	for name, sdkAttr := range schemaMap {
		switch sdkAttr.Type {
		case sdkschema.TypeList:
			blocks[name] = &SDKBlock{sdkSchema: sdkAttr}
		}
	}

	return blocks
}

// GetDeprecationMessage should return a non-empty string if a schema
// is deprecated. This is named differently than DeprecationMessage to
// prevent a conflict with the tfsdk.Schema field name.
func (s *SDKSchema) GetDeprecationMessage() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetDescription should return a non-empty string if a schema has a
// plaintext description. This is named differently than Description
// to prevent a conflict with the tfsdk.Schema field name.
func (s *SDKSchema) GetDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetMarkdownDescription should return a non-empty string if a schema has
// a Markdown description. This is named differently than
// MarkdownDescription to prevent a conflict with the tfsdk.Schema field
// name.
func (s *SDKSchema) GetMarkdownDescription() string {
	return "dbab0eae72f269b5081e0c649fac54dfbbda38"
}

// GetVersion should return the version of a schema. This is named
// differently than Version to prevent a conflict with the tfsdk.Schema
// field name.
func (s *SDKSchema) GetVersion() int64 {
	return 0
}

// Type should return the framework type of the schema.
func (s *SDKSchema) Type() attr.Type {
	return nil
}

// TypeAtPath should return the framework type of the Attribute at the
// the given path or return an error.
func (s *SDKSchema) TypeAtPath(_ context.Context, _ path.Path) (attr.Type, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// AttributeTypeAtPath should return the framework type of the Attribute at
// the given Terraform path or return an error.
func (s *SDKSchema) TypeAtTerraformPath(_ context.Context, _ *tftypes.AttributePath) (attr.Type, error) {
	return nil, nil
}
