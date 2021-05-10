package tf

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attribute"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestState(t *testing.T) {
	raw, err := tfprotov6.NewDynamicValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"my_string":          tftypes.String,
			"my_list_of_strings": tftypes.List{ElementType: tftypes.String},
		},
	}, tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"my_string":          tftypes.String,
			"id":                 tftypes.String,
			"my_list_of_strings": tftypes.List{ElementType: tftypes.String},
		},
	}, map[string]tftypes.Value{
		"my_string": tftypes.NewValue(tftypes.String, "katy"),
		"id":        tftypes.NewValue(tftypes.String, "static_id"),
		"my_list_of_strings": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "katy2"),
		}),
	}))

	schema := Schema{
		Attributes: map[string]Attribute{
			"my_string": {
				Type:      types.StringType{},
				Required:  true,
				Sensitive: true,
			},
			"my_list_of_strings": {
				Type:               types.ListType{ElemType: types.StringType{}},
				Optional:           true,
				Computed:           true,
				DeprecationMessage: "my_list_of_strings is deprecated and will be removed in the next major version. Please use my_nested_attributes instead.",
			},
			// "my_nested_attributes": {
			// 	Attributes: map[string]Attribute{
			// 		"my_nested_string": {
			// 			Type:     types.StringType{},
			// 			Optional: true,
			// 			Computed: true,
			// 		},
			// 	},
			// 	AttributesNestingMode: NestingModeList,
			// 	Optional:              true,
			// 	Computed:              true,
			// },
		},
	}

	state := State{
		Raw:    raw,
		Schema: schema,
	}

	// string test

	actualStr, err := state.GetString(context.Background(), tftypes.NewAttributePath().WithAttributeName("my_string"))
	if err != nil {
		t.Fatal(err)
	}

	expectedStr := types.String{
		Value: "katy",
	}

	if actualStr != expectedStr {
		t.Fatalf("expected %v, got %v", expectedStr, actualStr)
	}

	// list test

	actualList, err := state.Get(context.Background(), tftypes.NewAttributePath().WithAttributeName("my_list_of_strings"), types.ListType{ElemType: types.StringType{}})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("actualList %+v, elems %s", actualList, actualList.(types.List).Elems)

	expectedList := types.List{
		Elems:    []attribute.AttributeValue{types.String{Value: "katy2"}},
		ElemType: tftypes.String,
	}

	if expectedList.Equal(actualList) {
		t.Fatalf("expected %+v, got %+v", expectedList, actualList)
	}
}
