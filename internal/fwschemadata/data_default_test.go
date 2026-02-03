// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testdefaults"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataDefault(t *testing.T) {
	t.Parallel()

	var float32AttributeValue float32 = 1.2345
	var float32DefaultValue float32 = 5.4321

	testCases := map[string]struct {
		data          *fwschemadata.Data
		rawConfig     tftypes.Value
		expected      *fwschemadata.Data
		expectedDiags diag.Diagnostics
	}{
		"bool-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Bool{
								DefaultBoolMethod: func(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
									if !req.Path.Equal(path.Root("bool_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("bool_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Bool{
								DefaultBoolMethod: func(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
									if !req.Path.Equal(path.Root("bool_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("bool_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
		},
		"bool-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Bool{
								DefaultBoolMethod: func(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Bool{
								DefaultBoolMethod: func(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"bool-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, true), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"bool-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"bool-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
		},
		"bool-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool_attribute": tftypes.Bool,
				},
			},
				map[string]tftypes.Value{
					"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false),
					},
				),
			},
		},
		"float32-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float32{
								DefaultFloat32Method: func(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
									if !req.Path.Equal(path.Root("float32_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("float32_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float32{
								DefaultFloat32Method: func(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
									if !req.Path.Equal(path.Root("float32_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("float32_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
		},
		"float32-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float32{
								DefaultFloat32Method: func(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float32{
								DefaultFloat32Method: func(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"float32-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  float32default.StaticFloat32(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, 5.4321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  float32default.StaticFloat32(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float32-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float32Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float32Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float32-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  float32default.StaticFloat32(float32DefaultValue),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, float64(float32AttributeValue)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  float32default.StaticFloat32(float32DefaultValue),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, float64(float32DefaultValue)),
					},
				),
			},
		},
		"float32-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float32_attribute": testschema.AttributeWithFloat32DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float32_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float64-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float64{
								DefaultFloat64Method: func(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
									if !req.Path.Equal(path.Root("float64_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("float64_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float64{
								DefaultFloat64Method: func(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
									if !req.Path.Equal(path.Root("float64_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("float64_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
		},
		"float64-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float64{
								DefaultFloat64Method: func(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Float64{
								DefaultFloat64Method: func(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"float64-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, 5.4321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float64-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"float64-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  float64default.StaticFloat64(5.4321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 5.4321),
					},
				),
			},
		},
		"float64-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"float64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"float64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64_attribute": testschema.AttributeWithFloat64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64_attribute": tftypes.NewValue(tftypes.Number, 1.2345),
					},
				),
			},
		},
		"int32-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int32{
								DefaultInt32Method: func(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
									if !req.Path.Equal(path.Root("int32_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("int32_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int32{
								DefaultInt32Method: func(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
									if !req.Path.Equal(path.Root("int32_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("int32_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
		},
		"int32-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int32{
								DefaultInt32Method: func(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int32{
								DefaultInt32Method: func(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"int32-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  int32default.StaticInt32(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, 54321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  int32default.StaticInt32(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int32-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int32Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int32Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int32-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  int32default.StaticInt32(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  int32default.StaticInt32(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 54321),
					},
				),
			},
		},
		"int32-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int32_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int32_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int32_attribute": testschema.AttributeWithInt32DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int32_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int32_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int64-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int64{
								DefaultInt64Method: func(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
									if !req.Path.Equal(path.Root("int64_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("int64_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int64{
								DefaultInt64Method: func(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
									if !req.Path.Equal(path.Root("int64_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("int64_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
		},
		"int64-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int64{
								DefaultInt64Method: func(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Int64{
								DefaultInt64Method: func(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"int64-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, 54321), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int64-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"int64-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  int64default.StaticInt64(54321),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 54321),
					},
				),
			},
		},
		"int64-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"int64_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"int64_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64_attribute": testschema.AttributeWithInt64DefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64_attribute": tftypes.NewValue(tftypes.Number, 12345),
					},
				),
			},
		},
		"list-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.List{
								DefaultListMethod: func(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
									if !req.Path.Equal(path.Root("list_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("list_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"list_attribute": tftypes.List{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.List{
								DefaultListMethod: func(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
									if !req.Path.Equal(path.Root("list_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("list_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
					},
				),
			},
		},
		"list-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.List{
								DefaultListMethod: func(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"list_attribute": tftypes.List{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.List{
								DefaultListMethod: func(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"list-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"list-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"list-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"list-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									// intentionally incorrect element type
									types.BoolType,
									[]attr.Value{
										types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: listdefault.StaticValue(
								types.ListValueMust(
									// intentionally incorrect element type
									types.BoolType,
									[]attr.Value{
										types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"list_attribute\"): can't use tftypes.List[tftypes.Bool] as tftypes.List[tftypes.String]",
				),
			},
		},
		"list-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_attribute": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"list_attribute": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.AttributeWithListDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(tftypes.List{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Map{
								DefaultMapMethod: func(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
									if !req.Path.Equal(path.Root("map_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("map_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"map_attribute": tftypes.Map{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Map{
								DefaultMapMethod: func(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
									if !req.Path.Equal(path.Root("map_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("map_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					},
				),
			},
		},
		"map-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Map{
								DefaultMapMethod: func(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"map_attribute": tftypes.Map{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Map{
								DefaultMapMethod: func(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"map-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"map-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.StringType,
									map[string]attr.Value{
										"b": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"b": tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"map-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									// intentionally incorrect element type
									types.BoolType,
									map[string]attr.Value{
										"b": types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									// intentionally incorrect element type
									types.BoolType,
									map[string]attr.Value{
										"b": types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"map_attribute\"): can't use tftypes.Map[tftypes.Bool] as tftypes.Map[tftypes.String]",
				),
			},
		},
		"map-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_attribute": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"map_attribute": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map_attribute": testschema.AttributeWithMapDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_attribute": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map_attribute": tftypes.NewValue(tftypes.Map{
							ElementType: tftypes.String,
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"number-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Number{
								DefaultNumberMethod: func(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
									if !req.Path.Equal(path.Root("number_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("number_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Number{
								DefaultNumberMethod: func(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
									if !req.Path.Equal(path.Root("number_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("number_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
		},
		"number-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Number{
								DefaultNumberMethod: func(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Number{
								DefaultNumberMethod: func(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"number-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"number-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.NumberType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.NumberType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"number-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  numberdefault.StaticBigFloat(big.NewFloat(5.4321)),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(5.4321)),
					},
				),
			},
		},
		"number-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"number_attribute": tftypes.Number,
				},
			},
				map[string]tftypes.Value{
					"number_attribute": tftypes.NewValue(tftypes.Number, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"number_attribute": testschema.AttributeWithNumberDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"number_attribute": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"number_attribute": tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					},
				),
			},
		},
		"object-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional: true,
							Computed: true,
							AttributeTypes: map[string]attr.Type{
								"test_attribute": types.StringType,
							},
							Default: testdefaults.Object{
								DefaultObjectMethod: func(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
									if !req.Path.Equal(path.Root("object_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("object_attribute"), req.Path),
										)
									}

									// Response value type must conform to the schema or an error will be returned.
									resp.PlanValue = types.ObjectNull(
										map[string]attr.Type{
											"test_attribute": types.StringType,
										},
									)
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
				},
			},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional: true,
							Computed: true,
							AttributeTypes: map[string]attr.Type{
								"test_attribute": types.StringType,
							},
							Default: testdefaults.Object{
								DefaultObjectMethod: func(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
									if !req.Path.Equal(path.Root("object_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("object_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
					},
				),
			},
		},
		"object-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional: true,
							Computed: true,
							AttributeTypes: map[string]attr.Type{
								"test_attribute": types.StringType,
							},
							Default: testdefaults.Object{
								DefaultObjectMethod: func(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
				},
			},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional: true,
							Computed: true,
							AttributeTypes: map[string]attr.Type{
								"test_attribute": types.StringType,
							},
							Default: testdefaults.Object{
								DefaultObjectMethod: func(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"test_attribute": tftypes.String}}, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"object-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"object-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.Attribute{
							Computed: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"a": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.Attribute{
							Computed: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"a": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"object-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{"a": types.StringType},
									map[string]attr.Value{
										"a": types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"object-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									// intentionally invalid attribute types
									map[string]attr.Type{"invalid": types.BoolType},
									map[string]attr.Value{
										"invalid": types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									// intentionally invalid attribute types
									map[string]attr.Type{"invalid": types.BoolType},
									map[string]attr.Value{
										"invalid": types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"object_attribute\"): can't use tftypes.Object[\"invalid\":tftypes.Bool] as tftypes.Object[\"a\":tftypes.String]",
				),
			},
		},
		"object-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default:        nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"object_attribute": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						},
					},
				},
				map[string]tftypes.Value{
					"object_attribute": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object_attribute": testschema.AttributeWithObjectDefaultValue{
							Optional:       true,
							AttributeTypes: map[string]attr.Type{"a": types.StringType},
							Default:        nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
							},
						},
					},
					map[string]tftypes.Value{
						"object_attribute": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{"a": tftypes.String},
						}, map[string]tftypes.Value{
							"a": tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Set{
								DefaultSetMethod: func(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
									if !req.Path.Equal(path.Root("set_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("set_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"set_attribute": tftypes.Set{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Set{
								DefaultSetMethod: func(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
									if !req.Path.Equal(path.Root("set_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("set_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
					},
				),
			},
		},
		"set-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Set{
								DefaultSetMethod: func(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"set_attribute": tftypes.Set{ElementType: tftypes.String},
				},
			},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Default: testdefaults.Set{
								DefaultSetMethod: func(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{ElementType: tftypes.String},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"set-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
					}),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"set-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.StringType,
									[]attr.Value{
										types.StringValue("two"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "two"),
						}),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"set-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									// intentionally invalid element type
									types.BoolType,
									[]attr.Value{
										types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default: setdefault.StaticValue(
								types.SetValueMust(
									// intentionally invalid element type
									types.BoolType,
									[]attr.Value{
										types.BoolValue(true),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"set_attribute\"): can't use tftypes.Set[tftypes.Bool] as tftypes.Set[tftypes.String]",
				),
			},
		},
		"set-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_attribute": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"set_attribute": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.AttributeWithSetDefaultValue{
							Optional:    true,
							ElementType: types.StringType,
							Default:     nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(tftypes.Set{
							ElementType: tftypes.String,
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "one"),
						}),
					},
				),
			},
		},
		"string-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.String{
								DefaultStringMethod: func(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
									if !req.Path.Equal(path.Root("string_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("string_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.String{
								DefaultStringMethod: func(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
									if !req.Path.Equal(path.Root("string_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("string_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
		},
		"string-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.String{
								DefaultStringMethod: func(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.String{
								DefaultStringMethod: func(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, "two"), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  stringdefault.StaticString("two"),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "two"),
					},
				),
			},
		},
		"string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string_attribute": tftypes.String,
				},
			},
				map[string]tftypes.Value{
					"string_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string_attribute": testschema.AttributeWithStringDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string_attribute": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"dynamic-attribute-request-path": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Dynamic{
								DefaultDynamicMethod: func(ctx context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
									if !req.Path.Equal(path.Root("dynamic_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("dynamic_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Dynamic{
								DefaultDynamicMethod: func(ctx context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
									if !req.Path.Equal(path.Root("dynamic_attribute")) {
										resp.Diagnostics.AddError(
											"unexpected req.Path value",
											fmt.Sprintf("expected %s, got: %s", path.Root("dynamic_attribute"), req.Path),
										)
									}
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					},
				),
			},
		},
		"dynamic-attribute-response-diagnostics": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Dynamic{
								DefaultDynamicMethod: func(ctx context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Optional: true,
							Computed: true,
							Default: testdefaults.Dynamic{
								DefaultDynamicMethod: func(ctx context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
									resp.Diagnostics.AddError("test error summary", "test error detail")
									resp.Diagnostics.AddWarning("test warning summary", "test warning detail")
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("test error summary", "test error detail"),
				diag.NewWarningDiagnostic("test warning summary", "test warning detail"),
			},
		},
		"dynamic-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("two"))),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.String, "two"), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("two"))),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"dynamic-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								// Default transform walk will visit both of these elements and skip
								tftypes.NewValue(tftypes.String, "one"),
								tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "one"),
								tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
		},
		"dynamic-attribute-known-type-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig, type is known as String
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Computed: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"dynamic-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default: dynamicdefault.StaticValue(
								types.DynamicValue(
									types.ListValueMust(types.StringType, []attr.Value{
										types.StringValue("three"),
										types.StringValue("four"),
									}),
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								// Default transform walk will visit both of these elements and skip
								tftypes.NewValue(tftypes.String, "one"),
								tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("two"))),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "three"),
								tftypes.NewValue(tftypes.String, "four"),
							},
						),
					},
				),
			},
		},
		"dynamic-attribute-known-type-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("two"))),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig, type is known as String
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  dynamicdefault.StaticValue(types.DynamicValue(types.StringValue("two"))),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "two"),
					},
				),
			},
		},
		"dynamic-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.DynamicPseudoType, nil), // value in rawConfig
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"dynamic-attribute-known-type-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
			rawConfig: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"dynamic_attribute": tftypes.DynamicPseudoType,
				},
			},
				map[string]tftypes.Value{
					"dynamic_attribute": tftypes.NewValue(tftypes.String, nil), // value in rawConfig, type is known as String
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.AttributeWithDynamicDefaultValue{
							Computed: true,
							Default:  nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(tftypes.String, "one"),
					},
				),
			},
		},
		"list-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"list-nested-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									// intentionally invalid element type
									types.StringType,
									[]attr.Value{
										types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: listdefault.StaticValue(
								types.ListValueMust(
									// intentionally invalid element type
									types.StringType,
									[]attr.Value{
										types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"list_nested\"): can't use tftypes.List[tftypes.String] as tftypes.List[tftypes.Object[\"string_attribute\":tftypes.String]]",
				),
			},
		},
		"list-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": testschema.NestedAttributeWithListDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"list_nested": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"list_nested": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"list_nested": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									map[string]attr.Value{
										"test-key": types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"test-key": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									map[string]attr.Value{
										"test-key": types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									map[string]attr.Value{
										"test-key": types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									map[string]attr.Value{
										"test-key": types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"map-nested-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									// intentionally invalid element type
									types.StringType,
									map[string]attr.Value{
										"test-key": types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: mapdefault.StaticValue(
								types.MapValueMust(
									// intentionally invalid element type
									types.StringType,
									map[string]attr.Value{
										"test-key": types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"map_nested\"): can't use tftypes.Map[tftypes.String] as tftypes.Map[tftypes.Object[\"string_attribute\":tftypes.String]]",
				),
			},
		},
		"map-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": testschema.NestedAttributeWithMapDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"test-key": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"test-key": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"test-key": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"map-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"map_nested": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"map_nested": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						map[string]tftypes.Value{
							"test-key": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"map_nested": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map_nested": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map_nested": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"test-key": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"string_attribute": types.StringType,
										},
									},
									[]attr.Value{
										types.ObjectValueMust(
											map[string]attr.Type{
												"string_attribute": types.StringType,
											}, map[string]attr.Value{
												"string_attribute": types.StringValue("two"),
											}),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"set-nested-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									// intentionally invalid element type
									types.StringType,
									[]attr.Value{
										types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: setdefault.StaticValue(
								types.SetValueMust(
									// intentionally invalid element type
									types.StringType,
									[]attr.Value{
										types.StringValue("invalid"),
									},
								),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"set_nested\"): can't use tftypes.Set[tftypes.String] as tftypes.Set[tftypes.Object[\"string_attribute\":tftypes.String]]",
				),
			},
		},
		"set-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": testschema.NestedAttributeWithSetDefaultValue{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.Attribute{
										Computed: true,
										Type:     types.StringType,
									},
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, "one"),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  stringdefault.StaticString("two"),
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "two"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"set_nested": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"set_nested": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"string_attribute": tftypes.NewValue(tftypes.String, nil),
								},
							),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"set_nested": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": testschema.AttributeWithStringDefaultValue{
										Computed: true,
										Default:  nil,
									},
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"string_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"string_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"string_attribute": tftypes.NewValue(tftypes.String, "one"),
									},
								),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, "one"),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									map[string]attr.Type{
										"string_attribute": types.StringType,
									},
									map[string]attr.Value{
										"string_attribute": types.StringValue("two"),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/930
		"single-nested-attribute-null-invalid-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									// intentionally invalid attribute types
									map[string]attr.Type{
										"invalid": types.BoolType,
									},
									map[string]attr.Value{
										"invalid": types.BoolValue(true),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									// intentionally invalid attribute types
									map[string]attr.Type{
										"invalid": types.BoolType,
									},
									map[string]attr.Value{
										"invalid": types.BoolValue(true),
									}),
							),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Handling Schema Defaults",
					"An unexpected error occurred while handling schema default values. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"single_nested\"): can't use tftypes.Object[\"invalid\":tftypes.Bool] as tftypes.Object[\"string_attribute\":tftypes.String]",
				),
			},
		},
		"single-nested-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": testschema.NestedAttributeWithObjectDefaultValue{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
							Default: nil,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-not-null-unmodified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, "one"),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-unmodified-no-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-modified-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  stringdefault.StaticString("two"),
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "two"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-string-attribute-null-unmodified-default-nil": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  nil,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
			rawConfig: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"single_nested": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"single_nested": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string_attribute": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"string_attribute": tftypes.NewValue(tftypes.String, nil),
						},
					),
				},
			),
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"single_nested": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"string_attribute": testschema.AttributeWithStringDefaultValue{
									Computed: true,
									Default:  nil,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string_attribute": tftypes.NewValue(tftypes.String, "one"),
							},
						),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.TransformDefaults(context.Background(), testCase.rawConfig)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
