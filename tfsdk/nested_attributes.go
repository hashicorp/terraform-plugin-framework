package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

// ListNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration.
//
// Deprecated: Use datasource/schema.ListNestedAttribute,
// provider/schema.ListNestedAttribute, or resource/schema.ListNestedAttribute
// instead. This can be switched by using the datasource/schema.Schema,
// provider/schema.Schema, or resource/schema.Schema types.
func ListNestedAttributes(attributes map[string]Attribute) fwschema.NestedAttributes {
	return fwschema.ListNestedAttributes{
		UnderlyingAttributes: schemaAttributes(attributes),
	}
}

// MapNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration. Each group will need to be associated with a unique string by
// the user.
//
// Deprecated: Use datasource/schema.MapNestedAttribute,
// provider/schema.MapNestedAttribute, or resource/schema.MapNestedAttribute
// instead. This can be switched by using the datasource/schema.Schema,
// provider/schema.Schema, or resource/schema.Schema types.
func MapNestedAttributes(attributes map[string]Attribute) fwschema.NestedAttributes {
	return fwschema.MapNestedAttributes{
		UnderlyingAttributes: schemaAttributes(attributes),
	}
}

// SetNestedAttributes nests `attributes` under another attribute, allowing
// multiple instances of that group of attributes to appear in the
// configuration, while requiring each group of values be unique.
//
// Deprecated: Use datasource/schema.SetNestedAttribute,
// provider/schema.SetNestedAttribute, or resource/schema.SetNestedAttribute
// instead. This can be switched by using the datasource/schema.Schema,
// provider/schema.Schema, or resource/schema.Schema types.
func SetNestedAttributes(attributes map[string]Attribute) fwschema.NestedAttributes {
	return fwschema.SetNestedAttributes{
		UnderlyingAttributes: schemaAttributes(attributes),
	}
}

// SingleNestedAttributes nests `attributes` under another attribute, only
// allowing one instance of that group of attributes to appear in the
// configuration.
//
// Deprecated: Use datasource/schema.SingleNestedAttribute,
// provider/schema.SingleNestedAttribute, or resource/schema.SingleNestedAttribute
// instead. This can be switched by using the datasource/schema.Schema,
// provider/schema.Schema, or resource/schema.Schema types.
func SingleNestedAttributes(attributes map[string]Attribute) fwschema.NestedAttributes {
	return fwschema.SingleNestedAttributes{
		UnderlyingAttributes: schemaAttributes(attributes),
	}
}
