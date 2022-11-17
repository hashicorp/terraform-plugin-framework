package fwschema

// NestedAttribute defines a schema attribute that contains nested attributes.
type NestedAttribute interface {
	Attribute

	// GetAttributes should return the nested attributes of an attribute, if
	// applicable. This is named differently than Attribute to prevent a
	// conflict with the tfsdk.Attribute field name.
	GetAttributes() UnderlyingAttributes

	// GetNestingMode should return the nesting mode (list, map, set, or
	// single) of the nested attributes or left unset if this Attribute
	// does not represent nested attributes.
	GetNestingMode() NestingMode
}
