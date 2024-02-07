// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEqualityDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		"DynamicValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.DynamicValue(types.StringValue("prior")),
				ProposedNewValue: types.DynamicValue(types.StringValue("new")),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.DynamicValue(types.StringValue("new")),
			},
		},
		"DynamicValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("prior")),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("new")),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("prior")),
					SemanticEquals: true,
				},
			},
		},
		"DynamicValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("prior")),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("new")),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("new")),
					SemanticEquals: false,
				},
			},
		},
		"DynamicValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("prior")),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("new")),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.DynamicValueWithSemanticEquals{
					DynamicValue:   types.DynamicValue(types.StringValue("new")),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualityDynamic(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
