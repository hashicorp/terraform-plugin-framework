// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Type  = InvalidType{}
	_ attr.Value = Invalid{}
)

// InvalidType is an attr.Type that returns errors for methods that can return errors.
type InvalidType struct{}

func (t InvalidType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("intentional ApplyTerraform5AttributePathStep error")
}

func (t InvalidType) Equal(o attr.Type) bool {
	_, ok := o.(InvalidType)

	return ok
}

func (t InvalidType) String() string {
	return "testtypes.InvalidType"
}

func (t InvalidType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t InvalidType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	return nil, fmt.Errorf("intentional ValueFromTerraform error")
}

// ValueType returns the Value type.
func (t InvalidType) ValueType(_ context.Context) attr.Value {
	return Invalid{}
}

// Invalid is an attr.Value that returns errors for methods than can return errors.
type Invalid struct{}

func (i Invalid) Equal(o attr.Value) bool {
	_, ok := o.(Invalid)

	return ok
}

func (i Invalid) IsNull() bool {
	return false
}

func (i Invalid) IsUnknown() bool {
	return false
}

func (i Invalid) String() string {
	return "<invalid>"
}

func (i Invalid) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return tftypes.Value{}, fmt.Errorf("intentional ToTerraformValue error")
}

func (i Invalid) Type(_ context.Context) attr.Type {
	return InvalidType{}
}
