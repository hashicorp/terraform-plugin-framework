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
	// If Type is set, Attributes cannot be.
	Type attr.Type

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// If Attributes is set, Type cannot be.
	Attributes NestedAttributes

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
	Required bool

	// Optional indicates whether the practitioner can choose not to enter
	// a value for this attribute or not. Optional and Required cannot both
	// be true.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
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
// ApplyTerraform5AttributePathStep on a.Type or a.Attributes, whichever is
// non-nil. It allows Attributes to be walked using tftypes.Walk and
// tftypes.Transform.
func (a Attribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if a.Type != nil {
		return a.Type.ApplyTerraform5AttributePathStep(step)
	}
	if a.Attributes != nil {
		return a.Attributes.ApplyTerraform5AttributePathStep(step)
	}
	return nil, errors.New("Attribute has no type or nested attributes")
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

// tfprotov6 returns the *tfprotov6.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on
// `path`. `name` is the name of the attribute.
func (a Attribute) tfprotov6SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath) (*tfprotov6.SchemaAttribute, error) {
	schemaAttribute := &tfprotov6.SchemaAttribute{
		Name:      name,
		Required:  a.Required,
		Optional:  a.Optional,
		Computed:  a.Computed,
		Sensitive: a.Sensitive,
	}

	if !a.Required && !a.Optional && !a.Computed {
		return nil, path.NewErrorf("must have Required, Optional, or Computed set")
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

	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 && a.Type != nil {
		return nil, path.NewErrorf("can't have both Attributes and Type set")
	}

	if (a.Attributes == nil || len(a.Attributes.GetAttributes()) < 1) && a.Type == nil {
		return nil, path.NewErrorf("must have Attributes or Type set")
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
		nestedSchemaAttribute, err := nestedA.tfprotov6SchemaAttribute(ctx, nestedName, path.WithAttributeName(nestedName))

		if err != nil {
			return nil, err
		}

		object.Attributes = append(object.Attributes, nestedSchemaAttribute)
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

// validate performs all Attribute validation.
func (a Attribute) validate(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if (a.Attributes == nil || len(a.Attributes.GetAttributes()) == 0) && a.Type == nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute must define either Attributes or Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 && a.Type != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	if !a.Required && !a.Optional && !a.Computed {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Attribute Definition",
			"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
		)

		return
	}

	attributeConfig, diags := req.Config.GetAttribute(ctx, req.AttributePath)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	for _, validator := range a.Validators {
		validator.Validate(ctx, req, resp)
	}

	if a.Attributes != nil {
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

// modifyPlan runs all AttributePlanModifiers
func (a Attribute) modifyPlan(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	attrConfig, diags := req.Config.GetAttribute(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeConfig = attrConfig

	attrState, diags := req.State.GetAttribute(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)
	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeState = attrState

	attrPlan, diags := req.Plan.GetAttribute(ctx, req.AttributePath)
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
