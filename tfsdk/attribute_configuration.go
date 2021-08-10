package tfsdk

// AttributeConfiguration describes Attribute value behaviors.
type AttributeConfiguration int

const (
	// AttributeConfigurationUnknown signals an invalid Attribute definition.
	//
	// This is the default value of this type and is used to signal an
	// Attribute missing a proper Configuration definition.
	AttributeConfigurationUnknown AttributeConfiguration = 0

	// AttributeConfigurationRequired denotes a value that requires practitioner configuration.
	//
	// If the practitioner configuration does not include a known value, an
	// error is returned.
	AttributeConfigurationRequired AttributeConfiguration = 1

	// AttributeConfigurationOptional denotes a value that can be optionally set in a practitioner configuration.
	//
	// A plan difference for the value will be shown if plugin returns a value
	// and there is no practitioner configuration.
	AttributeConfigurationOptional AttributeConfiguration = 2

	// AttributeConfigurationOptionalComputed denotes a value that can be set either by practitioner configuration or the plugin.
	//
	// No plan difference for the value will be shown if plugin returns a value
	// and there is no practitioner configuration.
	AttributeConfigurationOptionalComputed AttributeConfiguration = 3

	// AttributeConfigurationComputed denotes a value that can only be set by the plugin.
	//
	// Effectively, this is a read-only Attribute.
	AttributeConfigurationComputed AttributeConfiguration = 4
)

// String returns a string representation of the AttributeConfiguration.
func (ac AttributeConfiguration) String() string {
	switch ac {
	case AttributeConfigurationRequired:
		return "Required"
	case AttributeConfigurationOptional:
		return "Optional"
	case AttributeConfigurationOptionalComputed:
		return "OptionalComputed"
	case AttributeConfigurationComputed:
		return "Computed"
	default:
		return "Unknown"
	}
}
