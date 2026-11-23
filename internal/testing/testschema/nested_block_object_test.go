// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testschema_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
)

func TestNestedBlockObjectRange(t *testing.T) {
	t.Parallel()

	attributes := map[string]fwschema.Attribute{
		"a": testschema.Attribute{Optional: true},
		"b": testschema.Attribute{Required: true},
	}
	blocks := map[string]fwschema.Block{
		"a": testschema.Block{NestingMode: fwschema.BlockNestingModeList},
		"b": testschema.Block{NestingMode: fwschema.BlockNestingModeSet},
	}
	testCases := map[string]fwschema.NestedBlockObject{
		"empty":               testschema.NestedBlockObject{},
		"base":                testschema.NestedBlockObject{Attributes: attributes, Blocks: blocks},
		"with-plan-modifiers": testschema.NestedBlockObjectWithPlanModifiers{Attributes: attributes, Blocks: blocks},
		"with-validators":     testschema.NestedBlockObjectWithValidators{Attributes: attributes, Blocks: blocks},
	}

	for name, object := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotAttributes := make(fwschema.UnderlyingAttributes)
			object.RangeAttributes(func(name string, attribute fwschema.Attribute) bool { gotAttributes[name] = attribute; return true })
			if want := object.GetAttributes(); len(want) != len(gotAttributes) || len(want) > 0 && !reflect.DeepEqual(want, gotAttributes) {
				t.Errorf("unexpected attributes: got %#v, want %#v", gotAttributes, object.GetAttributes())
			}
			gotBlocks := make(map[string]fwschema.Block)
			object.RangeBlocks(func(name string, block fwschema.Block) bool { gotBlocks[name] = block; return true })
			if want := object.GetBlocks(); len(want) != len(gotBlocks) || len(want) > 0 && !reflect.DeepEqual(want, gotBlocks) {
				t.Errorf("unexpected blocks: got %#v, want %#v", gotBlocks, object.GetBlocks())
			}
			attributeCount := 0
			object.RangeAttributes(func(string, fwschema.Attribute) bool { attributeCount++; return false })
			blockCount := 0
			object.RangeBlocks(func(string, fwschema.Block) bool { blockCount++; return false })
			if want := min(len(object.GetAttributes()), 1); attributeCount != want {
				t.Errorf("attribute early-stop count = %d, want %d", attributeCount, want)
			}
			if want := min(len(object.GetBlocks()), 1); blockCount != want {
				t.Errorf("block early-stop count = %d, want %d", blockCount, want)
			}
		})
	}
}
