// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// buildSetNested constructs a prior and proposed SetNested value of n object
// elements. When semanticEquals is false the object fields are plain string and
// int64 attributes (no semantic equality anywhere in the type tree). When true,
// the "name" field is a custom string type that always reports semantic
// equality, exercising the per-element reconciliation path.
func buildSetNested(ctx context.Context, tb testing.TB, n int, semanticEquals bool) (basetypes.SetValue, basetypes.SetValue) {
	tb.Helper()

	var nameType attr.Type = basetypes.StringType{}
	if semanticEquals {
		nameType = testtypes.StringTypeWithSemanticEquals{SemanticEquals: true}
	}

	attrTypes := map[string]attr.Type{
		"name": nameType,
		"num":  basetypes.Int64Type{},
	}
	objType := basetypes.ObjectType{AttrTypes: attrTypes}

	newName := func(s string) attr.Value {
		if semanticEquals {
			return testtypes.StringValueWithSemanticEquals{
				StringValue:    basetypes.NewStringValue(s),
				SemanticEquals: true,
			}
		}
		return basetypes.NewStringValue(s)
	}

	priorElements := make([]attr.Value, n)
	proposedElements := make([]attr.Value, n)

	for i := 0; i < n; i++ {
		priorObj, diags := basetypes.NewObjectValue(attrTypes, map[string]attr.Value{
			"name": newName(fmt.Sprintf("prior-%d", i)),
			"num":  basetypes.NewInt64Value(int64(i)),
		})
		if diags.HasError() {
			tb.Fatalf("prior object %d: %v", i, diags)
		}

		proposedObj, diags := basetypes.NewObjectValue(attrTypes, map[string]attr.Value{
			"name": newName(fmt.Sprintf("proposed-%d", i)),
			"num":  basetypes.NewInt64Value(int64(i)),
		})
		if diags.HasError() {
			tb.Fatalf("proposed object %d: %v", i, diags)
		}

		priorElements[i] = priorObj
		proposedElements[i] = proposedObj
	}

	prior, diags := basetypes.NewSetValue(objType, priorElements)
	if diags.HasError() {
		tb.Fatalf("prior set: %v", diags)
	}

	proposed, diags := basetypes.NewSetValue(objType, proposedElements)
	if diags.HasError() {
		tb.Fatalf("proposed set: %v", diags)
	}

	return prior, proposed
}

func benchmarkSetNested(b *testing.B, n int, semanticEquals bool) {
	ctx := context.Background()
	prior, proposed := buildSetNested(ctx, b, n, semanticEquals)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := fwschemadata.ValueSemanticEqualityRequest{
			Path:             path.Root("test"),
			PriorValue:       prior,
			ProposedNewValue: proposed,
		}
		resp := &fwschemadata.ValueSemanticEqualityResponse{
			NewValue: req.ProposedNewValue,
		}

		fwschemadata.ValueSemanticEquality(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			b.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
		}
	}
}

// No semantic equality anywhere in the type tree: this is the pathological case
// from the real-world report (e.g. a large cloudflare_list). The pre-fix code
// walks all N^2 element pairs here even though every recursive call is a no-op.
func BenchmarkSetNestedNoSemanticEquals100(b *testing.B)  { benchmarkSetNested(b, 100, false) }
func BenchmarkSetNestedNoSemanticEquals1000(b *testing.B) { benchmarkSetNested(b, 1000, false) }
func BenchmarkSetNestedNoSemanticEquals3251(b *testing.B) { benchmarkSetNested(b, 3251, false) }

// An element type implements semantic equality: the per-element walk must still
// run. Kept as a correctness/perf guard for the path the fix must preserve.
func BenchmarkSetNestedWithSemanticEquals100(b *testing.B) { benchmarkSetNested(b, 100, true) }

// TestBenchmarkSetNestedSemanticEqualsStillApplies proves the element-level
// semantic equality path still runs after the short-circuit was added: with a
// field type that always reports semantic equality, every proposed element must
// be swapped back to its prior counterpart, so the reconciled value differs
// from the proposed value and equals the prior value.
func TestBenchmarkSetNestedSemanticEqualsStillApplies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior, proposed := buildSetNested(ctx, t, 10, true)

	req := fwschemadata.ValueSemanticEqualityRequest{
		Path:             path.Root("test"),
		PriorValue:       prior,
		ProposedNewValue: proposed,
	}
	resp := &fwschemadata.ValueSemanticEqualityResponse{
		NewValue: req.ProposedNewValue,
	}

	fwschemadata.ValueSemanticEquality(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	if resp.NewValue.Equal(proposed) {
		t.Fatalf("expected reconciled value to differ from proposed value (semantic equality path did not run)")
	}

	if !resp.NewValue.Equal(prior) {
		t.Fatalf("expected reconciled value to equal prior value after semantic equality swap")
	}
}

// TestBenchmarkSetNestedNoSemanticEqualsUnchanged proves the short-circuit is
// behavior-preserving for the no-semantic-equality case: the reconciled value
// is byte-identical to the proposed value.
func TestBenchmarkSetNestedNoSemanticEqualsUnchanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior, proposed := buildSetNested(ctx, t, 10, false)

	req := fwschemadata.ValueSemanticEqualityRequest{
		Path:             path.Root("test"),
		PriorValue:       prior,
		ProposedNewValue: proposed,
	}
	resp := &fwschemadata.ValueSemanticEqualityResponse{
		NewValue: req.ProposedNewValue,
	}

	fwschemadata.ValueSemanticEquality(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	if !resp.NewValue.Equal(proposed) {
		t.Fatalf("expected reconciled value to equal proposed value when no element implements semantic equality")
	}
}
