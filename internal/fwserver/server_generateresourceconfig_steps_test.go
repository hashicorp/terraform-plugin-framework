// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testdefaults"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type testExactlyOneOfStringValidator struct {
	paths path.Expressions
}

func (v testExactlyOneOfStringValidator) Description(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) ValidateString(context.Context, validator.StringRequest, *validator.StringResponse) {
}

func (v testExactlyOneOfStringValidator) Paths() path.Expressions {
	return v.paths
}

type testAlsoRequiresStringValidator struct {
	paths path.Expressions
}

func (v testAlsoRequiresStringValidator) Description(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) ValidateString(context.Context, validator.StringRequest, *validator.StringResponse) {
}

func (v testAlsoRequiresStringValidator) Paths() path.Expressions {
	return v.paths
}

func TestResolveExactlyOneOfGroups(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"alpha": tftypes.String,
		"beta":  tftypes.String,
	}}

	testSchema := testschema.Schema{Attributes: map[string]fwschema.Attribute{
		"alpha": testschema.AttributeWithStringValidators{
			Computed: true,
			Optional: true,
			Validators: []validator.String{
				testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
			},
		},
		"beta": testschema.AttributeWithStringDefaultValue{
			Optional: true,
			Default: testdefaults.String{
				DefaultStringMethod: func(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
					resp.PlanValue = types.StringValue("beta-default")
				},
			},
		},
	}}

	testCases := map[string]struct {
		config   tftypes.Value
		expected tftypes.Value
	}{
		"sets default when all null": {
			config: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.String, nil),
				"beta":  tftypes.NewValue(tftypes.String, nil),
			}),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.String, nil),
				"beta":  tftypes.NewValue(tftypes.String, "beta-default"),
			}),
		},
		"keeps first non-null and nulls rest": {
			config: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
				"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
			}),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
				"beta":  tftypes.NewValue(tftypes.String, nil),
			}),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotDiags := resolveExactlyOneOfGroups(t.Context(), testCase.config, testSchema, diag.Diagnostics{})

			if diff := cmp.Diff(testCase.expected, got); diff != "" {
				t.Fatalf("unexpected config diff: %s", diff)
			}

			if len(gotDiags) != 0 {
				t.Fatalf("unexpected diagnostics: %v", gotDiags)
			}
		})
	}
}

func TestResolveAlsoRequiresGroups(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"alpha": tftypes.String,
		"beta":  tftypes.String,
		"gamma": tftypes.String,
	}}

	testSchema := testschema.Schema{Attributes: map[string]fwschema.Attribute{
		"alpha": testschema.AttributeWithStringValidators{
			Optional: true,
			Validators: []validator.String{
				testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
			},
		},
		"beta": testschema.AttributeWithStringValidators{
			Optional: true,
			Validators: []validator.String{
				testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("gamma")}},
			},
		},
		"gamma": testschema.Attribute{Optional: true, Type: types.StringType},
	}}

	config := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
		"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
		"gamma": tftypes.NewValue(tftypes.String, nil),
	})

	expected := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha": tftypes.NewValue(tftypes.String, nil),
		"beta":  tftypes.NewValue(tftypes.String, nil),
		"gamma": tftypes.NewValue(tftypes.String, nil),
	})

	got, gotDiags := resolveAlsoRequiresGroups(t.Context(), config, testSchema, diag.Diagnostics{})

	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatalf("unexpected config diff: %s", diff)
	}

	if len(gotDiags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", gotDiags)
	}
}

var (
	_ validator.String                = testExactlyOneOfStringValidator{}
	_ validator.ExactlyOneOfValidator = testExactlyOneOfStringValidator{}
	_ validator.String                = testAlsoRequiresStringValidator{}
	_ validator.AlsoRequiresValidator = testAlsoRequiresStringValidator{}
)
