package types

import (
	"context"
	"fmt"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type primitive uint8

const (
	// StringType represents a UTF-8 string type.
	StringType primitive = 1

	// NumberType represents a number type, either an integer or a float.
	NumberType primitive = 2

	// BoolType represents a boolean type.
	BoolType primitive = 3
)

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (p primitive) TerraformType(_ context.Context) tftypes.Type {
	switch p {
	case StringType:
		return tftypes.String
	case NumberType:
		return tftypes.Number
	case BoolType:
		return tftypes.Bool
	default:
		panic(fmt.Sprintf("unknown primitive %d", p))
	}
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (p primitive) ValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	switch p {
	case StringType:
		return stringValueFromTerraform(ctx, in)
	case NumberType:
		return numberValueFromTerraform(ctx, in)
	case BoolType:
		return boolValueFromTerraform(ctx, in)
	default:
		panic(fmt.Sprintf("unknown primitive %d", p))
	}
}

func (p primitive) Equal(o tfsdk.AttributeType) bool {
	other, ok := o.(primitive)
	if !ok {
		return false
	}
	return p == other
}
