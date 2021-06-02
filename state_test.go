package tfsdk

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStateGet(t *testing.T) {
	schema := schema.Schema{
		Attributes: map[string]schema.Attribute{
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
		},
	}
	state := State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"foo": tftypes.String,
				"bar": tftypes.List{ElementType: tftypes.String},
			},
		}, map[string]tftypes.Value{
			"foo": tftypes.NewValue(tftypes.String, "hello, world"),
			"bar": tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "red"),
				tftypes.NewValue(tftypes.String, "blue"),
				tftypes.NewValue(tftypes.String, "green"),
			}),
		}),
		Schema: schema,
	}
	type myType struct {
		Foo types.String `tfsdk:"foo"`
		Bar types.List   `tfsdk:"bar"`
	}
	var val myType
	err := state.Get(context.Background(), &val)
	if err != nil {
		t.Errorf("Error running As: %s", err)
	}
	if val.Foo.Unknown {
		t.Error("Expected Foo to be known")
	}
	if val.Foo.Null {
		t.Error("Expected Foo to be non-null")
	}
	if val.Foo.Value != "hello, world" {
		t.Errorf("Expected Foo to be %q, got %q", "hello, world", val.Foo.Value)
	}
	if val.Bar.Unknown {
		t.Error("Expected Bar to be known")
	}
	if val.Bar.Null {
		t.Errorf("Expected Bar to be non-null")
	}
	if len(val.Bar.Elems) != 3 {
		t.Errorf("Expected Bar to have 3 elements, had %d", len(val.Bar.Elems))
	}
	if val.Bar.Elems[0].(types.String).Value != "red" {
		t.Errorf("Expected Bar's first element to be %q, got %q", "red", val.Bar.Elems[0].(types.String).Value)
	}
	if val.Bar.Elems[1].(types.String).Value != "blue" {
		t.Errorf("Expected Bar's second element to be %q, got %q", "blue", val.Bar.Elems[1].(types.String).Value)
	}
	if val.Bar.Elems[2].(types.String).Value != "green" {
		t.Errorf("Expected Bar's third element to be %q, got %q", "green", val.Bar.Elems[2].(types.String).Value)
	}
}
