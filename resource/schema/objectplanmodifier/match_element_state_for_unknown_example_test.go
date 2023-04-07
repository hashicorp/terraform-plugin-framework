package objectplanmodifier_test

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func ExampleMatchElementStateForUnknown() {
	// Used within a Schema method of a Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_attr": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attr": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"nested_computed_attr": schema.StringAttribute{
									Computed: true,
								},
							},
							Computed: true,
							PlanModifiers: []planmodifier.Object{
								// Preseve this computed value during updates.
								objectplanmodifier.MatchElementStateForUnknown(
									// Identify matching prior state value based on configurable_attr
									path.MatchRelative().AtParent().AtName("configurable_attr"),
									// ... potentially others ...
								),
							},
						},
						"configurable_attr": schema.StringAttribute{
							Required: true,
						},
					},
				},
				Optional: true,
			},
		},
	}
}
