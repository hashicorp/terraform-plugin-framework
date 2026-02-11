// Copyright IBM Corp. 2021, 2026
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

func TestValueSemanticEqualityInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		"Int32Value": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.Int32Value(12),
				ProposedNewValue: types.Int32Value(24),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.Int32Value(24),
			},
		},
		"Int32ValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(12),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(24),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(12),
					SemanticEquals: true,
				},
			},
		},
		"Int32ValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(12),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(24),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(24),
					SemanticEquals: false,
				},
			},
		},
		"Int32ValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(12),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(24),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int32ValueWithSemanticEquals{
					Int32Value:     types.Int32Value(24),
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualityInt32(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
