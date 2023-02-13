package fwxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

// AttributeWithBoolDefaultValue is an optional interface on Attribute which
// enables Bool default value support.
type AttributeWithBoolDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Bool
}

// AttributeWithFloat64DefaultValue is an optional interface on Attribute which
// enables Float64 default value support.
type AttributeWithFloat64DefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Float64
}

// AttributeWithInt64DefaultValue is an optional interface on Attribute which
// enables Int64 default value support.
type AttributeWithInt64DefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Int64
}

// AttributeWithListDefaultValue is an optional interface on Attribute which
// enables List default value support.
type AttributeWithListDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.List
}

// AttributeWithMapDefaultValue is an optional interface on Attribute which
// enables Map default value support.
type AttributeWithMapDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Map
}

// AttributeWithNumberDefaultValue is an optional interface on Attribute which
// enables Number default value support.
type AttributeWithNumberDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Number
}

// AttributeWithObjectDefaultValue is an optional interface on Attribute which
// enables Object default value support.
type AttributeWithObjectDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Object
}

// AttributeWithSetDefaultValue is an optional interface on Attribute which
// enables Set default value support.
type AttributeWithSetDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.Set
}

// AttributeWithStringDefaultValue is an optional interface on Attribute which
// enables String default value support.
type AttributeWithStringDefaultValue interface {
	fwschema.Attribute

	DefaultValue() defaults.String
}
