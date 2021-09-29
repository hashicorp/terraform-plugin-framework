package tfsdk

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Attribute defines the constraints and behaviors of a single field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
	// Type indicates what kind of attribute this is. You'll most likely
	// want to use one of the types in the types package.
	//
	// If Type is set, Attributes and Blocks cannot be.
	Type attr.Type

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// In practitioner configurations, an equals sign (=) is required to set
	// the value. See also:
	//   https://www.terraform.io/docs/language/syntax/configuration.html
	//
	// If Attributes is set, Blocks and Type cannot be. Attributes are strongly
	// preferred over Blocks.
	Attributes NestedAttributes

	// Blocks can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
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
	// If Blocks is set, Attributes, Computed, Optional, Required, Sensitive,
	// and Type cannot be. Attributes are strongly preferred over Blocks.
	// Blocks should only be used for configuration compatibility with
	// previously existing schemas from an older Terraform Plugin SDK. Efforts
	// should be made to convert Blocks to Attributes as a breaking change for
	// practitioners.
	Blocks NestedBlocks

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	//
	// Cannot be set with Blocks. Only nested attributes under Blocks may set
	// configurability. Use ListNestedBlockOptions with MinItems above 0 to
	// mark enforce Block requirement in a practitioner configuration.
	Required bool

	// Optional indicates whether the practitioner can choose not to enter
	// a value for this attribute or not. Optional and Required cannot both
	// be true.
	//
	// Cannot be set with Blocks. Only nested attributes under Blocks may set
	// configurability.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	//
	// Cannot be set with Blocks. Only nested attributes under Blocks may set
	// configurability.
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
	//
	// Cannot be set with Blocks. Only nested attributes under Blocks may set
	// sensitivity.
	Sensitive bool

	// DeprecationMessage defines a message to display to practitioners
	// using this attribute, warning them that it is deprecated and
	// instructing them on what upgrade steps to take.
	DeprecationMessage string

	// Validators defines validation functionality for the attribute.
	Validators []AttributeValidator

	// PlanModifiers defines a sequence of modifiers for this attribute at
	// plan time. Attribute-level plan modifications occur before any
	// resource-level plan modifications.
	//
	// Any errors will prevent further execution of this sequence
	// of modifiers and modifiers associated with any nested Attribute, but will not
	// prevent execution of PlanModifiers on any other Attribute in the Schema.
	//
	// Plan modification only applies to resources, not data sources or
	// providers. Setting PlanModifiers on a data source or provider attribute
	// will have no effect.
	PlanModifiers AttributePlanModifiers
}

// ApplyTerraform5AttributePathStep transparently calls
// ApplyTerraform5AttributePathStep on a.Type, a.Attributes or a.Blocks,
// whichever is non-nil. It allows Attributes to be walked using tftypes.Walk
// and tftypes.Transform.
func (a Attribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if a.Type != nil {
		return a.Type.ApplyTerraform5AttributePathStep(step)
	}
	if a.Attributes != nil {
		return a.Attributes.ApplyTerraform5AttributePathStep(step)
	}
	if a.Blocks != nil {
		return a.Blocks.ApplyTerraform5AttributePathStep(step)
	}
	return nil, errors.New("Attribute has no type, nested attributes, or nested blocks")
}

// Equal returns true if `a` and `o` should be considered Equal.
func (a Attribute) Equal(o Attribute) bool {
	if a.Type == nil && o.Type != nil {
		return false
	} else if a.Type != nil && o.Type == nil {
		return false
	} else if a.Type != nil && o.Type != nil && !a.Type.Equal(o.Type) {
		return false
	}
	if a.Attributes == nil && o.Attributes != nil {
		return false
	} else if a.Attributes != nil && o.Attributes == nil {
		return false
	} else if a.Attributes != nil && o.Attributes != nil && !a.Attributes.Equal(o.Attributes) {
		return false
	}
	if a.Blocks == nil && o.Blocks != nil {
		return false
	} else if a.Blocks != nil && o.Blocks == nil {
		return false
	} else if a.Blocks != nil && o.Blocks != nil && !a.Blocks.Equal(o.Blocks) {
		return false
	}
	if a.Description != o.Description {
		return false
	}
	if a.MarkdownDescription != o.MarkdownDescription {
		return false
	}
	if a.Required != o.Required {
		return false
	}
	if a.Optional != o.Optional {
		return false
	}
	if a.Computed != o.Computed {
		return false
	}
	if a.Sensitive != o.Sensitive {
		return false
	}
	if a.DeprecationMessage != o.DeprecationMessage {
		return false
	}
	return true
}

// definesAttributes returns true if Attribute has a non-empty Blocks definition.
//
// Attribute may also incorrectly have an Attributes and/or Type definition.
func (a Attribute) definesAttributes() bool {
	return a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0
}

// definesBlocks returns true if Attribute has a non-empty Blocks definition.
//
// Attribute may also incorrectly have an Attributes and/or Type definition.
func (a Attribute) definesBlocks() bool {
	return a.Blocks != nil && len(a.Blocks.GetAttributes()) > 0
}

// tfprotov6 returns the *tfprotov6.SchemaAttribute or
// *tfprotov6.SchemaNestedBlock equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on
// `path`. `name` is the name of the attribute.
func (a Attribute) tfprotov6(ctx context.Context, name string, path *tftypes.AttributePath) (interface{}, error) {
	if !a.definesAttributes() && !a.definesBlocks() && a.Type == nil {
		return nil, path.NewErrorf("must have Attributes, Blocks, or Type set")
	}

	if a.definesBlocks() {
		return a.tfprotov6SchemaNestedBlock(ctx, name, path)
	}

	return a.tfprotov6SchemaAttribute(ctx, name, path)
}

// tfprotov6SchemaAttribute returns the *tfprotov6.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on
// `path`. `name` is the name of the attribute.
func (a Attribute) tfprotov6SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath) (*tfprotov6.SchemaAttribute, error) {
	if a.definesAttributes() && a.definesBlocks() {
		return nil, path.NewErrorf("cannot have both Attributes and Blocks set")
	}

	if a.definesAttributes() && a.Type != nil {
		return nil, path.NewErrorf("cannot have both Attributes and Type set")
	}

	if !a.definesAttributes() && a.Type == nil {
		return nil, path.NewErrorf("must have Attributes or Type set")
	}

	if !a.Required && !a.Optional && !a.Computed {
		return nil, path.NewErrorf("must have Required, Optional, or Computed set")
	}

	schemaAttribute := &tfprotov6.SchemaAttribute{
		Name:      name,
		Required:  a.Required,
		Optional:  a.Optional,
		Computed:  a.Computed,
		Sensitive: a.Sensitive,
	}

	if a.DeprecationMessage != "" {
		schemaAttribute.Deprecated = true
	}

	if a.Description != "" {
		schemaAttribute.Description = a.Description
		schemaAttribute.DescriptionKind = tfprotov6.StringKindPlain
	}

	if a.MarkdownDescription != "" {
		schemaAttribute.Description = a.MarkdownDescription
		schemaAttribute.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	if a.Type != nil {
		schemaAttribute.Type = a.Type.TerraformType(ctx)

		return schemaAttribute, nil
	}

	object := &tfprotov6.SchemaObject{
		MinItems: a.Attributes.GetMinItems(),
		MaxItems: a.Attributes.GetMaxItems(),
	}
	nm := a.Attributes.GetNestingMode()
	switch nm {
	case NestingModeSingle:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSingle
	case NestingModeList:
		object.Nesting = tfprotov6.SchemaObjectNestingModeList
	case NestingModeSet:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSet
	case NestingModeMap:
		object.Nesting = tfprotov6.SchemaObjectNestingModeMap
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for nestedName, nestedA := range a.Attributes.GetAttributes() {
		nestedPath := path.WithAttributeName(nestedName)
		nestedAProto6Raw, err := nestedA.tfprotov6(ctx, nestedName, nestedPath)

		if err != nil {
			return nil, err
		}

		switch nestedAProto6 := nestedAProto6Raw.(type) {
		case *tfprotov6.SchemaAttribute:
			object.Attributes = append(object.Attributes, nestedAProto6)
		case *tfprotov6.SchemaNestedBlock:
			return nil, nestedPath.NewErrorf("cannot have Blocks inside Attributes")
		default:
			return nil, nestedPath.NewErrorf("unknown tfprotov6 type %T in Attributes", nestedAProto6Raw)
		}
	}

	sort.Slice(object.Attributes, func(i, j int) bool {
		if object.Attributes[i] == nil {
			return true
		}

		if object.Attributes[j] == nil {
			return false
		}

		return object.Attributes[i].Name < object.Attributes[j].Name
	})

	schemaAttribute.NestedType = object

	return schemaAttribute, nil
}

// tfprotov6SchemaNestedBlock returns the *tfprotov6.SchemaNestedBlock
// equivalent of an Attribute. Errors will be tftypes.AttributePathErrors based
// on `path`. `name` is the name of the attribute.
func (a Attribute) tfprotov6SchemaNestedBlock(ctx context.Context, name string, path *tftypes.AttributePath) (*tfprotov6.SchemaNestedBlock, error) {
	if a.definesAttributes() {
		return nil, path.NewErrorf("cannot have both Attributes and Blocks set")
	}

	if a.Computed {
		return nil, path.NewErrorf("cannot set Block as Computed, mark all nested Attributes instead")
	}

	if a.Optional {
		return nil, path.NewErrorf("cannot set Block as Optional, mark all nested Attributes instead")
	}

	if a.Required {
		return nil, path.NewErrorf("cannot set Block as Required, mark all nested Attributes instead")
	}

	if a.Sensitive {
		return nil, path.NewErrorf("cannot set Block as Sensitive, mark all nested Attributes instead")
	}

	if a.Type != nil {
		return nil, path.NewErrorf("cannot have both Blocks and Type set")
	}

	schemaNestedBlock := &tfprotov6.SchemaNestedBlock{
		Block: &tfprotov6.SchemaBlock{
			Deprecated: a.DeprecationMessage != "",
		},
		MinItems: a.Blocks.GetMinItems(),
		MaxItems: a.Blocks.GetMaxItems(),
		TypeName: name,
	}

	if a.Description != "" {
		schemaNestedBlock.Block.Description = a.Description
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindPlain
	}

	if a.MarkdownDescription != "" {
		schemaNestedBlock.Block.Description = a.MarkdownDescription
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	nm := a.Blocks.GetNestingMode()
	switch nm {
	case NestingModeList:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeList
	case NestingModeSet:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeSet
	case NestingModeMap, NestingModeSingle:
		// This is intentional to only maintain the previous Terraform Plugin SDK support.
		return nil, path.NewErrorf("unsupported Blocks nesting mode: %v", nm)
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for nestedName, nestedA := range a.Blocks.GetAttributes() {
		nestedPath := path.WithAttributeName(nestedName)
		nestedAProto6Raw, err := nestedA.tfprotov6(ctx, nestedName, nestedPath)

		if err != nil {
			return nil, err
		}

		switch nestedAProto6 := nestedAProto6Raw.(type) {
		case *tfprotov6.SchemaAttribute:
			schemaNestedBlock.Block.Attributes = append(schemaNestedBlock.Block.Attributes, nestedAProto6)
		case *tfprotov6.SchemaNestedBlock:
			schemaNestedBlock.Block.BlockTypes = append(schemaNestedBlock.Block.BlockTypes, nestedAProto6)
		default:
			return nil, nestedPath.NewErrorf("unknown tfprotov6 type %T in Blocks", nestedAProto6Raw)
		}
	}

	sort.Slice(schemaNestedBlock.Block.Attributes, func(i, j int) bool {
		if schemaNestedBlock.Block.Attributes[i] == nil {
			return true
		}

		if schemaNestedBlock.Block.Attributes[j] == nil {
			return false
		}

		return schemaNestedBlock.Block.Attributes[i].Name < schemaNestedBlock.Block.Attributes[j].Name
	})

	sort.Slice(schemaNestedBlock.Block.BlockTypes, func(i, j int) bool {
		if schemaNestedBlock.Block.BlockTypes[i] == nil {
			return true
		}

		if schemaNestedBlock.Block.BlockTypes[j] == nil {
			return false
		}

		return schemaNestedBlock.Block.BlockTypes[i].TypeName < schemaNestedBlock.Block.BlockTypes[j].TypeName
	})

	return schemaNestedBlock, nil
}

// validate performs all Attribute validation.
func (a Attribute) validate(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if !a.definesAttributes() && !a.definesBlocks() && a.Type == nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute must define either Attributes, Blocks, or Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if a.definesAttributes() && a.definesBlocks() {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Attributes and Blocks. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if a.definesAttributes() && a.Type != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if a.definesBlocks() && a.Type != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Blocks and Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if !a.definesBlocks() && !a.Required && !a.Optional && !a.Computed {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	attributeConfig, diags := req.Config.getAttributeValue(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	for _, validator := range a.Validators {
		validator.Validate(ctx, req, resp)
	}

	a.validateAttributes(ctx, req, resp)
	a.validateBlocks(ctx, req, resp)

	if a.DeprecationMessage != "" && attributeConfig != nil {
		tfValue, err := attributeConfig.ToTerraformValue(ctx)

		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot convert value. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if tfValue != nil {
			resp.Diagnostics.AddAttributeWarning(
				req.AttributePath,
				"Attribute Deprecated",
				a.DeprecationMessage,
			)
		}
	}
}

// validateAttributes performs all nested Attributes validation.
func (a Attribute) validateAttributes(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if !a.definesAttributes() {
		return
	}

	nm := a.Attributes.GetNestingMode()
	switch nm {
	case NestingModeList:
		l, ok := req.AttributeConfig.(types.List)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeSet:
		s, ok := req.AttributeConfig.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValueRaw, err := value.ToTerraformValue(ctx)

			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Attribute Validation Error",
					"Attribute validation cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			tfValue := tftypes.NewValue(s.ElemType.TerraformType(ctx), tfValueRaw)

			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeMap:
		m, ok := req.AttributeConfig.(types.Map)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for key := range m.Elems {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyString(key).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeSingle:
		o, ok := req.AttributeConfig.(types.Object)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if !o.Null && !o.Unknown {
			for nestedName, nestedAttr := range a.Attributes.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	default:
		err := fmt.Errorf("unknown attribute validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Attribute Validation Error",
			"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}
}

// validateBlocks performs all nested Blocks validation.
func (a Attribute) validateBlocks(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if !a.definesBlocks() {
		return
	}

	nm := a.Blocks.GetNestingMode()
	switch nm {
	case NestingModeList:
		l, ok := req.AttributeConfig.(types.List)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for nestedName, nestedAttr := range a.Blocks.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeSet:
		s, ok := req.AttributeConfig.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Validation Error",
				"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValueRaw, err := value.ToTerraformValue(ctx)

			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Attribute Validation Error",
					"Attribute validation cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			tfValue := tftypes.NewValue(s.ElemType.TerraformType(ctx), tfValueRaw)

			for nestedName, nestedAttr := range a.Blocks.GetAttributes() {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(nestedName),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				nestedAttr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	default:
		err := fmt.Errorf("unknown attribute validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Attribute Validation Error",
			"Attribute validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}
}

// modifyPlan runs all AttributePlanModifiers
func (a Attribute) modifyPlan(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrConfig, diags := req.Config.getAttributeValue(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeConfig = attrConfig

	attrState, diags := req.State.getAttributeValue(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeState = attrState

	attrPlan, diags := req.Plan.getAttributeValue(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributePlan = attrPlan

	for _, planModifier := range a.PlanModifiers {
		modifyResp := &ModifyAttributePlanResponse{
			AttributePlan:   resp.AttributePlan,
			RequiresReplace: resp.RequiresReplace,
		}

		planModifier.Modify(ctx, req, modifyResp)

		req.AttributePlan = modifyResp.AttributePlan
		resp.AttributePlan = modifyResp.AttributePlan
		resp.Diagnostics.Append(modifyResp.Diagnostics...)
		resp.RequiresReplace = modifyResp.RequiresReplace

		// Only on new errors.
		if modifyResp.Diagnostics.HasError() {
			return
		}
	}
}
