package fwserver

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"
)

func (s *Server) GenerateResourceConfigFuncs(ctx context.Context) (map[string]func() resource.ConfigModifiers, diag.Diagnostics) {
	provider, ok := s.Provider.(provider.ProviderWithGenerateResourceConfig)
	if !ok {
		return nil, nil
	}

	logging.FrameworkTrace(ctx, "Checking GenerateResourceConfigFuncs lock")
	s.listResourceFuncsMutex.Lock()
	defer s.listResourceFuncsMutex.Unlock()

	if s.generateResourceConfigFuncs != nil {
		return s.generateResourceConfigFuncs, s.generateResourceConfigFuncsDiags
	}

	s.generateResourceConfigFuncs = make(map[string]func() resource.ConfigModifiers)

	logging.FrameworkTrace(ctx, "Calling provider defined GenerateResourceConfigs")
	configModifiersFuncSlice := provider.GenerateResourceConfigs(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined GenerateResourceConfigs")

	for _, configModifierFunc := range configModifiersFuncSlice {
		configModifier := configModifierFunc()

		metadataReq := resource.MetadataRequest{
			ProviderTypeName: s.ProviderTypeName(ctx),
		}
		metadataResp := resource.MetadataResponse{}

		configModifier.Metadata(ctx, metadataReq, &metadataResp)

		typeName := metadataResp.TypeName
		if typeName == "" {
			// TODO error
			continue
		}

		logging.FrameworkTrace(ctx, "Found ConfigModifiers for type name: %s", map[string]interface{}{logging.KeyListResourceType: typeName})

		if _, ok := s.generateResourceConfigFuncs[typeName]; ok {
			// TODO error
			continue
		}

		s.generateResourceConfigFuncs[typeName] = configModifierFunc

	}

	return s.generateResourceConfigFuncs, s.generateResourceConfigFuncsDiags
}

func (s *Server) ModifyConfig(ctx context.Context, req resource.ModifyConfigRequest, resp *resource.ModifyConfigResponse) {
	s.generateResourceConfigFuncsMutex.Lock()
	modifyConfigFunc, ok := s.generateResourceConfigFuncs[req.TypeName]
	s.generateResourceConfigFuncsMutex.Unlock()

	if ok {
		modifyConfigFunc().ModifyConfig(ctx, req, resp)
	}

	return
}

type GenerateResourceConfigRequest struct {
	TypeName string
	State    *tfsdk.State
}

type GenerateResourceConfigResponse struct {
	Config      *tfsdk.Config
	Diagnostics diag.Diagnostics
}

func (s *Server) GenerateResourceConfig(ctx context.Context, req *GenerateResourceConfigRequest, resp *GenerateResourceConfigResponse) {
	if req == nil {
		return
	}

	if req.State == nil {
		// TODO error
	}

	stateData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionState,
		Schema:         req.State.Schema,
		TerraformValue: req.State.Raw,
	}

	configObj := make(map[string]tftypes.Type, 0)
	configAttributes := make(map[string]tftypes.Value, 0)
	schemaAttributes := make(map[string]resourceSchema.Attribute, 0)

	for name, attribute := range req.State.Schema.GetAttributes() {
		// skip all computed-only attributes
		if attribute.IsComputed() && !attribute.IsOptional() {
			// TODO these should be nil'ed?
			continue
		}

		//if v, ok := attribute.(fwschema.AttributeWithBoolDefaultValue); ok {
		//
		//}

		var tfVal tftypes.Value

		// Other cases to handle/consider:
		// AtLeastOneOf
		// ConflictsWith
		// Int Between Validations

		val, diags := stateData.ValueAtPath(ctx, path.Root(name))

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		// TODO handle the error
		tfVal, err := val.ToTerraformValue(ctx)
		if err != nil {
			fmt.Print(err)
		}
		//resp.Diagnostics.Append(err...)
		//if diags.HasError() {
		//	return
		//}

		if strings.EqualFold(name, "location") {
			tfVal = tftypes.NewValue(tftypes.String, nil)
		}

		if strings.EqualFold(name, "flow_timeout_in_minutes") {
			tfVal = tftypes.NewValue(tftypes.Number, nil)
		}

		configAttributes[name] = tfVal
		configObj[name] = tfVal.Type()

		schemaAttributes[name] = attribute
	}

	config := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{AttributeTypes: configObj}, configAttributes),
		Schema: &resourceSchema.Schema{
			Attributes: schemaAttributes,
			// TODO blocks
		},
	}

	cmReq := resource.ModifyConfigRequest{
		TypeName: req.TypeName,
		Config:   config,
	}

	cmResp := resource.ModifyConfigResponse{
		Config: config,
	}

	s.ModifyConfig(ctx, cmReq, &cmResp)

	resp.Diagnostics.Append(cmResp.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Config = &cmResp.Config
}
