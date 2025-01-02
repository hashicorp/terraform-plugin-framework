// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwtype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestContainsAllWriteOnlyChildAttributes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		nestedAttr metaschema.NestedAttribute
		expected   bool
	}{
		"empty nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{},
			expected:   false,
		},
		"writeOnly list nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
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
		"writeOnly list nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
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
		"writeOnly list nested attribute with multiple writeOnly child attributes returns true": {
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
		"writeOnly list nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
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
			expected: false,
		},
		"writeOnly list nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
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
		"writeOnly list nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"list_nested_attribute": schema.ListNestedAttribute{
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
		"writeOnly list nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
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
		"writeOnly list nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
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
		"non-writeOnly list nested attribute with one non-writeOnly child attribute returns false": {
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
		"non-writeOnly list nested attribute with writeOnly child nested attributes returns true": {
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
			expected: false,
		},
		"writeOnly set nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
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
		"writeOnly set nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
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
		"writeOnly set nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
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
		"writeOnly set nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
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
			expected: false,
		},
		"writeOnly set nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
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
		"writeOnly set nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
		"writeOnly set nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
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
		"writeOnly set nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
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
		"non-writeOnly set nested attribute with one non-writeOnly child attribute returns false": {
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
		"non-writeOnly set nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
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
			expected: false,
		},
		"writeOnly map nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
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
		"writeOnly map nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
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
		"writeOnly map nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
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
		"writeOnly map nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
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
			expected: false,
		},
		"writeOnly map nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
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
		"writeOnly map nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"map_nested_attribute": schema.MapNestedAttribute{
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
		"writeOnly map nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
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
		"writeOnly map nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
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
		"non-writeOnly map nested attribute with one non-writeOnly child attribute returns false": {
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
		"non-writeOnly map nested attribute with writeOnly child nested attributes returns true": {
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
			expected: false,
		},
		"writeOnly single nested attribute with writeOnly child attribute returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: true,
					},
				},
			},
			expected: true,
		},
		"writeOnly single nested attribute with non-writeOnly child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						WriteOnly: false,
					},
				},
			},
			expected: false,
		},
		"writeOnly single nested attribute with multiple writeOnly child attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
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
			expected: true,
		},
		"writeOnly single nested attribute with one non-writeOnly child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
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
			expected: false,
		},
		"writeOnly single nested attribute with writeOnly child nested attributes returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
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
		"writeOnly single nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
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
		"writeOnly single nested attribute with one non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
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
		"writeOnly single nested attribute with one non-writeOnly nested child attribute returns false": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
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
		"non-writeOnly single nested attribute with one non-writeOnly child attribute returns false": {
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
		"non-writeOnly single nested attribute with writeOnly child nested attributes returns true": {
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
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := fwtype.ContainsAllWriteOnlyChildAttributes(tt.nestedAttr); got != tt.expected {
				t.Errorf("ContainsAllWriteOnlyChildAttributes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContainsAnyWriteOnlyChildAttributes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		nestedAttr metaschema.NestedAttribute
		expected   bool
	}{
		"empty nested attribute returns false": {
			nestedAttr: schema.ListNestedAttribute{},
			expected:   false,
		},
		"list nested attribute with writeOnly returns true": {
			nestedAttr: schema.ListNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			expected: true,
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
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
		"set nested attribute with writeOnly returns true": {
			nestedAttr: schema.SetNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Computed: true,
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
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with non-writeOnly child nested attribute returns false": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		"set nested attribute with one non-writeOnly child nested attribute returns true": {
			nestedAttr: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"set_nested_attribute": schema.SetNestedAttribute{
							WriteOnly: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		"set nested attribute with one non-writeOnly nested child attribute returns true": {
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
		"map nested attribute with writeOnly returns true": {
			nestedAttr: schema.MapNestedAttribute{
				WriteOnly: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string_attribute": schema.StringAttribute{
							Computed: true,
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
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
									},
								},
							},
						},
						"list_nested_attribute": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"string_attribute": schema.StringAttribute{
										Computed: true,
									},
									"float32_attribute": schema.Float32Attribute{
										Computed: true,
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
		"single nested attribute with writeOnly returns true": {
			nestedAttr: schema.SingleNestedAttribute{
				WriteOnly: true,
				Attributes: map[string]schema.Attribute{
					"string_attribute": schema.StringAttribute{
						Computed: true,
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
								Computed: true,
							},
							"float32_attribute": schema.Float32Attribute{
								Computed: true,
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
						Attributes: map[string]schema.Attribute{
							"string_attribute": schema.StringAttribute{
								Computed: true,
							},
							"float32_attribute": schema.Float32Attribute{
								Computed: true,
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
								Computed: true,
							},
							"float32_attribute": schema.Float32Attribute{
								Computed: true,
							},
						},
					},
					"list_nested_attribute": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"string_attribute": schema.StringAttribute{
									Computed: true,
								},
								"float32_attribute": schema.Float32Attribute{
									Computed: true,
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
			if got := fwtype.ContainsAnyWriteOnlyChildAttributes(tt.nestedAttr); got != tt.expected {
				t.Errorf("ContainsAllWriteOnlyChildAttributes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
