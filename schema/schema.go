package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema is used to define the shape of practitioner-provider information,
// like resources, data sources, and providers. Think of it as a type
// definition, but for Terraform.
type Schema struct {
	// Attributes are the fields inside the resource, provider, or data
	// source that the schema is defining. The map key should be the name
	// of the attribute, and the body defines how it behaves. Names must
	// only contain lowercase letters, numbers, and underscores.
	Attributes map[string]Attribute

	// Version indicates the current version of the schema. Schemas are
	// versioned to help with automatic upgrade process. This is not
	// typically required unless there is a change in the schema, such as
	// changing an attribute type, that needs manual upgrade handling.
	// Versions should only be incremented by one each release.
	Version int64
}

func (s Schema) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if v, ok := step.(tftypes.AttributeName); ok {
		if attr, ok := s.Attributes[string(v)]; ok {
			return attr, nil
		} else {
			return nil, fmt.Errorf("could not find attribute %q in schema", v)
		}
	} else {
		return nil, fmt.Errorf("cannot apply AttributePathStep %T to schema", step)
	}
}
