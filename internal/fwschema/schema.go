package fwschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema is the core interface required for data sources, providers, and
// resources.
type Schema interface {
	// Implementations should include the tftypes.AttributePathStepper
	// interface methods for proper path and data handling.
	tftypes.AttributePathStepper

	// AttributeAtPath should return the Attribute at the given path or return
	// an error. This signature matches the existing tfsdk.Schema type
	// AttributeAtPath method signature to prevent the need for a breaking
	// change or deprecation of that method to create this interface.
	AttributeAtPath(path *tftypes.AttributePath) (Attribute, error)

	// AttributeType should return the framework type of the schema. This is
	// named differently than the Attribute interface GetType method name to
	// match the existing tfsdk.Schema type AttributeType method signature and
	// to prevent the need for a breaking change or deprecation of that method
	// to create this interface.
	AttributeType() attr.Type

	// AttributeTypeAtPath should return the framework type of the Attribute at
	// the given path or return an error. This signature matches the existing
	// tfsdk.Schema type AttributeAtPath method signature to prevent the need
	// for a breaking change or deprecation of that method to create this
	// interface.
	//
	// This will likely be removed in the future in preference of
	// AttributeAtPath.
	AttributeTypeAtPath(path *tftypes.AttributePath) (attr.Type, error)

	// GetAttributes should return the attributes of a schema. This is named
	// differently than Attributes to prevent a conflict with the tfsdk.Schema
	// field name.
	GetAttributes() map[string]Attribute

	// GetBlocks should return the blocks of a schema. This is named
	// differently than Blocks to prevent a conflict with the tfsdk.Schema
	// field name.
	GetBlocks() map[string]Block

	// GetDeprecationMessage should return a non-empty string if a schema
	// is deprecated. This is named differently than DeprecationMessage to
	// prevent a conflict with the tfsdk.Schema field name.
	GetDeprecationMessage() string

	// GetDescription should return a non-empty string if a schema has a
	// plaintext description. This is named differently than Description
	// to prevent a conflict with the tfsdk.Schema field name.
	GetDescription() string

	// GetMarkdownDescription should return a non-empty string if a schema has
	// a Markdown description. This is named differently than
	// MarkdownDescription to prevent a conflict with the tfsdk.Schema field
	// name.
	GetMarkdownDescription() string

	// GetVersion should return the version of a schema. This is named
	// differently than Version to prevent a conflict with the tfsdk.Schema
	// field name.
	GetVersion() int64

	// TerraformType should return the Terraform type of the schema. This
	// signature matches the existing tfsdk.Schema type TerraformType method
	// signature to prevent the need for a breaking change or deprecation of
	// that method to create this interface.
	//
	// This will likely be removed in the future in preferene of AttributeType.
	TerraformType(ctx context.Context) tftypes.Type
}
