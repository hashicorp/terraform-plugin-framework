package tfsdk

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaAttributeType(t *testing.T) {
	testSchema := Schema{
		Attributes: map[string]Attribute{
			"foo": {
				Type:     types.StringType,
				Required: true,
			},
			"bar": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"disks": {
				Attributes: ListNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type:     types.BoolType,
						Optional: true,
					},
				}, ListNestedAttributesOptions{}),
				Optional: true,
				Computed: true,
			},
			"boot_disk": {
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type: types.BoolType,
					},
				}),
			},
		},
	}

	expectedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"foo": types.StringType,
			"bar": types.ListType{
				ElemType: types.StringType,
			},
			"disks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"delete_with_instance": types.BoolType,
					},
				},
			},
			"boot_disk": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
			},
		},
	}

	actualType := testSchema.AttributeType()

	if !expectedType.Equal(actualType) {
		t.Fatalf("types not equal (+wanted, -got): %s", cmp.Diff(expectedType, actualType))
	}
}
