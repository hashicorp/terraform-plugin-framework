// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package path

import "testing"

func TestExpressionCopyIndependent(t *testing.T) {
	t.Parallel()

	original := Expression{root: true, steps: ExpressionSteps{
		ExpressionStepAttributeNameExact("first"),
		ExpressionStepAttributeNameExact("second"),
	}}
	copied := original.Copy()

	copied.steps[0] = ExpressionStepAttributeNameExact("copy")
	if got, want := original.steps[0], ExpressionStepAttributeNameExact("first"); got != want {
		t.Errorf("modifying copy changed original: got %v, want %v", got, want)
	}

	original.steps[1] = ExpressionStepAttributeNameExact("original")
	if got, want := copied.steps[1], ExpressionStepAttributeNameExact("second"); got != want {
		t.Errorf("modifying original changed copy: got %v, want %v", got, want)
	}
}
