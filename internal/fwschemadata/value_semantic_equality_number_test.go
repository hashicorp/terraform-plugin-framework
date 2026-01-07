// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEqualityNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		"NumberValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.NumberValue(big.NewFloat(1.2)),
				ProposedNewValue: types.NumberValue(big.NewFloat(2.4)),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.NumberValue(big.NewFloat(2.4)),
			},
		},
		"NumberValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: true,
				},
			},
		},
		"NumberValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
				},
			},
		},
		"NumberValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
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

			fwschemadata.ValueSemanticEqualityNumber(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
