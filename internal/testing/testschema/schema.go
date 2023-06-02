// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwschema.Schema = Schema{}

type Schema struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Version             int64
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Schema interface.
func (s Schema) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.SchemaApplyTerraform5AttributePathStep(s, step)
}

// AttributeAtPath satisfies the fwschema.Schema interface.
func (s Schema) AttributeAtPath(ctx context.Context, p path.Path) (fwschema.Attribute, diag.Diagnostics) {
	return fwschema.SchemaAttributeAtPath(ctx, s, p)
}

// AttributeAtTerraformPath satisfies the fwschema.Schema interface.
func (s Schema) AttributeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (fwschema.Attribute, error) {
	return fwschema.SchemaAttributeAtTerraformPath(ctx, s, p)
}

// GetAttributes satisfies the fwschema.Schema interface.
func (s Schema) GetAttributes() map[string]fwschema.Attribute {
	return s.Attributes
}

// GetBlocks satisfies the fwschema.Schema interface.
func (s Schema) GetBlocks() map[string]fwschema.Block {
	return s.Blocks
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

// Type satisfies the fwschema.Schema interface.
func (s Schema) Type() attr.Type {
	return fwschema.SchemaType(s)
}

// TypeAtPath satisfies the fwschema.Schema interface.
func (s Schema) TypeAtPath(ctx context.Context, p path.Path) (attr.Type, diag.Diagnostics) {
	return fwschema.SchemaTypeAtPath(ctx, s, p)
}

// TypeAtTerraformPath satisfies the fwschema.Schema interface.
func (s Schema) TypeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (attr.Type, error) {
	return fwschema.SchemaTypeAtTerraformPath(ctx, s, p)
}
