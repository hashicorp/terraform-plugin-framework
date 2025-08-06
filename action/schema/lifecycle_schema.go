// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ SchemaType = LifecycleSchema{}

// LifecycleSchema defines the structure and value types of a lifecycle action. A lifecycle action
// can cause changes to exactly one resource state, defined as a linked resource.
type LifecycleSchema struct {
	// ExecutionOrder defines when the lifecycle action must be executed in relation to the linked resource,
	// either before or after the linked resource's plan/apply.
	ExecutionOrder ExecutionOrder

	// LinkedResource represents the managed resource type that this action can make state changes to. The linked
	// resource must be defined in the same provider as the action is defined.
	//
	//  - If the managed resource is built with terraform-plugin-framework, use [LinkedResource].
	//  - If the managed resource is built with terraform-plugin-sdk/v2 or the terraform-plugin-go tfprotov5 package, use [RawV5LinkedResource].
	//  - If the managed resource is built with the terraform-plugin-go tfprotov6 package, use [RawV6LinkedResource].
	//
	// As a lifecycle action can only have a single linked resource, this linked resource data will always be at index 0
	// in the ModifyPlan and Invoke LinkedResources slice.
	LinkedResource LinkedResourceType

	// Attributes is the mapping of underlying attribute names to attribute
	// definitions.
	//
	// Names must only contain lowercase letters, numbers, and underscores.
	// Names must not collide with any Blocks names.
	Attributes map[string]Attribute

	// Blocks is the mapping of underlying block names to block definitions.
	//
	// Names must only contain lowercase letters, numbers, and underscores.
	// Names must not collide with any Attributes names.
	Blocks map[string]Block

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this action is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this action is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// DeprecationMessage defines warning diagnostic details to display when
	// practitioner configurations use this action. The warning diagnostic
	// summary is automatically set to "Action Deprecated" along with
	// configuration source file and line information.
	//
	// Set this field to a practitioner actionable message such as:
	//
	//  - "Use examplecloud_do_thing action instead. This action
	//    will be removed in the next major version of the provider."
	//  - "Remove this action as it no longer is valid and
	//    will be removed in the next major version of the provider."
	//
	DeprecationMessage string
}

func (s LifecycleSchema) LinkedResourceTypes() []LinkedResourceType {
	return []LinkedResourceType{
		s.LinkedResource,
	}
}

func (s LifecycleSchema) isActionSchemaType() {}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// schema.
func (s LifecycleSchema) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.SchemaApplyTerraform5AttributePathStep(s, step)
}

// AttributeAtPath returns the Attribute at the passed path. If the path points
// to an element or attribute of a complex type, rather than to an Attribute,
// it will return an ErrPathInsideAtomicAttribute error.
func (s LifecycleSchema) AttributeAtPath(ctx context.Context, p path.Path) (fwschema.Attribute, diag.Diagnostics) {
	return fwschema.SchemaAttributeAtPath(ctx, s, p)
}

// AttributeAtPath returns the Attribute at the passed path. If the path points
// to an element or attribute of a complex type, rather than to an Attribute,
// it will return an ErrPathInsideAtomicAttribute error.
func (s LifecycleSchema) AttributeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (fwschema.Attribute, error) {
	return fwschema.SchemaAttributeAtTerraformPath(ctx, s, p)
}

// GetAttributes returns the Attributes field value.
func (s LifecycleSchema) GetAttributes() map[string]fwschema.Attribute {
	return schemaAttributes(s.Attributes)
}

// GetBlocks returns the Blocks field value.
func (s LifecycleSchema) GetBlocks() map[string]fwschema.Block {
	return schemaBlocks(s.Blocks)
}

// GetDeprecationMessage returns the DeprecationMessage field value.
func (s LifecycleSchema) GetDeprecationMessage() string {
	return s.DeprecationMessage
}

// GetDescription returns the Description field value.
func (s LifecycleSchema) GetDescription() string {
	return s.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (s LifecycleSchema) GetMarkdownDescription() string {
	return s.MarkdownDescription
}

// GetVersion always returns 0 as action schemas cannot be versioned.
func (s LifecycleSchema) GetVersion() int64 {
	return 0
}

// Type returns the framework type of the schema.
func (s LifecycleSchema) Type() attr.Type {
	return fwschema.SchemaType(s)
}

// TypeAtPath returns the framework type at the given schema path.
func (s LifecycleSchema) TypeAtPath(ctx context.Context, p path.Path) (attr.Type, diag.Diagnostics) {
	return fwschema.SchemaTypeAtPath(ctx, s, p)
}

// TypeAtTerraformPath returns the framework type at the given tftypes path.
func (s LifecycleSchema) TypeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (attr.Type, error) {
	return fwschema.SchemaTypeAtTerraformPath(ctx, s, p)
}

// ValidateImplementation contains logic for validating the provider-defined
// implementation of the schema and underlying attributes and blocks to prevent
// unexpected errors or panics. This logic runs during the GetProviderSchema RPC,
// or via provider-defined unit testing, and should never include false positives.
func (s LifecycleSchema) ValidateImplementation(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	// TODO:Actions: Implement validation to ensure valid lifecycle "execute" enum and linked resource definitions

	for attributeName, attribute := range s.GetAttributes() {
		req := fwschema.ValidateImplementationRequest{
			Name: attributeName,
			Path: path.Root(attributeName),
		}

		// TODO:Actions: We should confirm with core, but we should be able to remove this next line.
		//
		// Action schemas define a specific "config" nested block in the action block, which means there
		// shouldn't be any conflict with existing or future Terraform core attributes.
		diags.Append(fwschema.IsReservedResourceAttributeName(req.Name, req.Path)...)
		diags.Append(fwschema.ValidateAttributeImplementation(ctx, attribute, req)...)
	}

	for blockName, block := range s.GetBlocks() {
		req := fwschema.ValidateImplementationRequest{
			Name: blockName,
			Path: path.Root(blockName),
		}

		// TODO:Actions: We should confirm with core, but we should be able to remove this next line.
		//
		// Action schemas define a specific "config" nested block in the action block, which means there
		// shouldn't be any conflict with existing or future Terraform core attributes.
		diags.Append(fwschema.IsReservedResourceAttributeName(req.Name, req.Path)...)
		diags.Append(fwschema.ValidateBlockImplementation(ctx, block, req)...)
	}

	return diags
}
