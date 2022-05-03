package proto6server

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// SchemaBlockAtPath returns the Block at the passed path. If the path points
// to an element or attribute of a complex type, rather than to a Block,
// it will return an ErrPathInsideAtomicAttribute error.
//
// TODO: Clean up this abstraction back into an internal Schema type method.
// The extra Schema parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func SchemaBlockAtPath(s tfsdk.Schema, path *tftypes.AttributePath) (tfsdk.Block, error) {
	res, remaining, err := tftypes.WalkAttributePath(s, path)
	if err != nil {
		return tfsdk.Block{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}

	switch r := res.(type) {
	// TODO: Temporarily not checked while this is only used in testing.
	// case nestedBlock:
	// 	return Block{}, ErrPathInsideAtomicAttribute
	case tfsdk.Block:
		return r, nil
	default:
		return tfsdk.Block{}, fmt.Errorf("got unexpected type %T", res)
	}
}
