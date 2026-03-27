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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetDefaultValueAtPathAppliesStringDefault(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"alpha": tftypes.String,
	}}

	testSchema := testschema.Schema{Attributes: map[string]fwschema.Attribute{
		"alpha": testschema.AttributeWithStringDefaultValue{
			Optional: true,
			Default: testdefaults.String{DefaultStringMethod: func(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
				resp.PlanValue = types.StringValue("alpha-default")
			}},
		},
	}}

	config := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha": tftypes.NewValue(tftypes.String, nil),
	})

	got, applied, gotDiags := setDefaultValueAtPath(t.Context(), config, testSchema, path.Root("alpha"), diag.Diagnostics{})
	if !applied {
		t.Fatal("expected default to be applied")
	}
	if len(gotDiags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", gotDiags)
	}

	expected := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha": tftypes.NewValue(tftypes.String, "alpha-default"),
	})
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatalf("unexpected config diff: %s", diff)
	}
}
