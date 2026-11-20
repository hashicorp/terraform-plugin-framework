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

func TestExpressionMergeIndependent(t *testing.T) {
	t.Parallel()

	for _, otherRoot := range []bool{false, true} {
		t.Run(map[bool]string{false: "relative", true: "root"}[otherRoot], func(t *testing.T) {
			t.Parallel()

			original := Expression{root: true, steps: ExpressionSteps{ExpressionStepAttributeNameExact("original")}}
			other := Expression{root: otherRoot, steps: ExpressionSteps{ExpressionStepAttributeNameExact("other")}}
			merged := original.Merge(other)
			if !merged.root {
				t.Error("merge did not preserve root")
			}
			otherIndex := 0
			if !otherRoot {
				otherIndex = 1
			}

			merged.steps[otherIndex] = ExpressionStepAttributeNameExact("changed-merged")
			if got, want := other.steps[0], ExpressionStepAttributeNameExact("other"); got != want {
				t.Errorf("modifying result changed other: got %v, want %v", got, want)
			}
			if !otherRoot {
				merged.steps[0] = ExpressionStepAttributeNameExact("changed-merged-original")
				if got, want := original.steps[0], ExpressionStepAttributeNameExact("original"); got != want {
					t.Errorf("modifying result changed original: got %v, want %v", got, want)
				}
			}

			other.steps[0] = ExpressionStepAttributeNameExact("changed-other")
			if got, want := merged.steps[otherIndex], ExpressionStepAttributeNameExact("changed-merged"); got != want {
				t.Errorf("modifying other changed result: got %v, want %v", got, want)
			}
			if !otherRoot {
				original.steps[0] = ExpressionStepAttributeNameExact("changed-original")
				if got, want := merged.steps[0], ExpressionStepAttributeNameExact("changed-merged-original"); got != want {
					t.Errorf("modifying original changed result: got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestExpressionMergeEmpty(t *testing.T) {
	t.Parallel()

	merged := (Expression{}).Merge(Expression{})
	if merged.root || merged.steps == nil || len(merged.steps) != 0 {
		t.Errorf("unexpected empty merge: %#v", merged)
	}

	relative := (Expression{}).Merge(Expression{steps: ExpressionSteps{ExpressionStepParent{}}})
	if relative.root || len(relative.steps) != 1 {
		t.Errorf("unexpected relative merge: %#v", relative)
	}
}
