package mapplanmodifier_test

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExampleMatchElementStateForUnknown() {
	// Used within a Schema method of a Resource
	_ = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"example_attr": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"computed_attr": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Map{
								// Preseve this computed value during updates.
								mapplanmodifier.MatchElementStateForUnknown(
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
