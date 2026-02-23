// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestContainsAllWriteOnlyChildAttributes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		nestedAttr schema.NestedAttribute
		expected   bool
	}{
		"empty nested attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{},
			expected:   true,
		},
		"list nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"set nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"set nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"single nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"single nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: false,
					},
				},
			},
			expected: false,
		},
		"single nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"single nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: false,
					},
				},
			},
			expected: false,
		},
		"single nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: true,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: false,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: false,
		},
		"single nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: true,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: true,
							},
						},
					},
					"list_nested_attribute": schema.ListNestedAttribute{
						WriteOnly: false,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									WriteOnly: true,
								},
								"float32_attribute": schema.Float32Attribute{
									WriteOnly: true,
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"single nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: true,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
							},
						},
					},
				},
			},
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fwschema.ContainsAllWriteOnlyChildAttributes(tt.nestedAttr); got != tt.expected {
				t.Errorf("ContainsAllWriteOnlyChildAttributes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContainsAnyWriteOnlyChildAttributes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		nestedAttr schema.NestedAttribute
		expected   bool
	}{
		"empty nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{},
			expected:   false,
		},
		"list nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with one non-writeOnly child attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"list nested attribute with one non-writeOnly child nested attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list nested attribute with one non-writeOnly nested child attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"set nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with one non-writeOnly child attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
										Optional:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
										Optional:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with no writeOnly child nested attributes returns false": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"set nested attribute with one writeOnly nested child attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with one non-writeOnly child attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"map nested attribute with one non-writeOnly child nested attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"map nested attribute with one non-writeOnly nested child attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},

		"single nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"single nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: false,
					},
				},
			},
			expected: false,
		},
		"single nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"single nested attribute with one non-writeOnly child attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: false,
					},
				},
			},
			expected: true,
		},
		"single nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: true,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
								Computed:  true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
								Computed:  true,
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: false,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
								Computed:  true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
								Computed:  true,
							},
						},
					},
				},
			},
			expected: false,
		},
		"single nested attribute with one non-writeOnly child nested attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: true,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
								Computed:  true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
								Computed:  true,
							},
						},
					},
					"list_nested_attribute": schema.ListNestedAttribute{
						WriteOnly: false,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									WriteOnly: false,
									Computed:  true,
								},
								"float32_attribute": schema.Float32Attribute{
									WriteOnly: false,
									Computed:  true,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested attribute with one non-writeOnly nested child attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						WriteOnly: false,
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fwschema.ContainsAnyWriteOnlyChildAttributes(tt.nestedAttr); got != tt.expected {
				t.Errorf("ContainsAllWriteOnlyChildAttributes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBlockContainsAnyWriteOnlyChildAttributes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		block    schema.Block
		expected bool
	}{
		"empty nested block returns false": {
			block:    schema.ListNestedBlock{},
			expected: false,
		},
		"list nested block with writeOnly child attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested block with non-writeOnly child attribute returns false": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"list nested block with multiple writeOnly child attributes returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"list nested block with one non-writeOnly child attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: true,
		},
		"list nested block with writeOnly child nested attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list nested block with non-writeOnly child nested attribute returns false": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"list nested block with one non-writeOnly child nested attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list double-nested block with top-level writeOnly nested child attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
					Blocks: map[string]schema.Block{
						"double_list_nested_block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"list double-nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: false,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
					Blocks: map[string]schema.Block{
						"double_list_nested_block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"list_nested_attribute": schema.ListNestedAttribute{
										WriteOnly: false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"string_attribute": schema.StringAttribute{
													WriteOnly: false,
												},
												"float32_attribute": schema.Float32Attribute{
													WriteOnly: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested block with non-writeOnly child attribute returns false": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: false,
		},
		"set nested block with multiple writeOnly child attributes returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: true,
						},
					},
				},
			},
			expected: true,
		},
		"set nested block with one non-writeOnly child attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							WriteOnly: true,
						},
						"float32_attribute": schema.Float32Attribute{
							WriteOnly: false,
						},
					},
				},
			},
			expected: true,
		},
		"set nested block with writeOnly child nested attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested block with non-writeOnly child nested attribute returns false": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"set nested block with one non-writeOnly child nested attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
										Computed:  true,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
										Computed:  true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set double-nested block with top-level writeOnly nested child attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
					Blocks: map[string]schema.Block{
						"double_set_nested_block": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set double-nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: false,
									},
								},
							},
						},
					},
					Blocks: map[string]schema.Block{
						"double_set_nested_block": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"set_nested_attribute": schema.SetNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"string_attribute": schema.StringAttribute{
													WriteOnly: false,
												},
												"float32_attribute": schema.Float32Attribute{
													WriteOnly: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested block with non-writeOnly child attribute returns false": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: false,
					},
				},
			},
			expected: false,
		},
		"single nested block with multiple writeOnly child attributes returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"single nested block with one non-writeOnly child attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
					"float32_attribute": schema.Float32Attribute{
						WriteOnly: false,
					},
				},
			},
			expected: true,
		},
		"single nested block with writeOnly child nested attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						WriteOnly: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									WriteOnly: false,
									Computed:  true,
								},
								"float32_attribute": schema.Float32Attribute{
									WriteOnly: false,
									Computed:  true,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested block with non-writeOnly child nested attribute returns false": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
								Computed:  true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
								Computed:  true,
							},
						},
					},
				},
			},
			expected: false,
		},
		"single nested block with one non-writeOnly child nested attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
								Computed:  true,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
								Computed:  true,
							},
						},
					},
					"list_nested_attribute": schema.ListNestedAttribute{
						WriteOnly: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									WriteOnly: false,
									Computed:  true,
								},
								"float32_attribute": schema.Float32Attribute{
									WriteOnly: false,
									Computed:  true,
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"single nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: true,
		},
		"single double-nested block with top-level writeOnly nested child attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"double_single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: true,
		},
		"single double-nested block with one non-writeOnly nested child attribute returns true": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								WriteOnly: false,
							},
							"float32_attribute": schema.Float32Attribute{
								WriteOnly: false,
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"double_single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"single_nested_attribute": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										WriteOnly: false,
									},
									"float32_attribute": schema.Float32Attribute{
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fwschema.BlockContainsAnyWriteOnlyChildAttributes(tt.block); got != tt.expected {
				t.Errorf("BlockContainsAllWriteOnlyChildAttributes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
