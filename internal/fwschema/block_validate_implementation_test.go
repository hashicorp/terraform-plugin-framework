// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschema_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var benchmarkBlockValidationDiagnostics diag.Diagnostics

func BenchmarkValidateBlockImplementation(b *testing.B) {
	for _, count := range []int{0, 1, 4, 16} {
		attributes := make(map[string]schema.Attribute, count)
		for i := range count {
			attributes[fmt.Sprintf("attribute_%d", i)] = schema.StringAttribute{Optional: true}
		}
		var blocks map[string]schema.Block
		for range 4 {
			blocks = map[string]schema.Block{"nested": schema.ListNestedBlock{NestedObject: schema.NestedBlockObject{Attributes: attributes, Blocks: blocks}}}
		}
		block := blocks["nested"]
		request := fwschema.ValidateImplementationRequest{Name: "root", Path: path.Root("root")}
		b.Run(fmt.Sprintf("attributes-%d", count), func(b *testing.B) {
			for b.Loop() {
				benchmarkBlockValidationDiagnostics = fwschema.ValidateBlockImplementation(context.Background(), block, request)
				if benchmarkBlockValidationDiagnostics.HasError() {
					b.Fatalf("unexpected diagnostics: %v", benchmarkBlockValidationDiagnostics)
				}
			}
		})
	}
}
