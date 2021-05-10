package tf

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attribute"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type State struct {
	Raw    tfprotov6.DynamicValue
	Schema Schema
}

// makeAttributeTypesObject converts a Schema into a tftypes map object used
// for unmarshalling the raw state.
func makeAttributeTypesObject(ctx context.Context, schema Schema) tftypes.Object {
	ret := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}

	for k, v := range schema.Attributes {
		if v.Type != nil {
			ret.AttributeTypes[k] = v.Type.TerraformType(ctx)
		}
	}

	return ret
}

// GetString attempts to return the state value at attributePath as a
// types.String.
func (s *State) GetString(ctx context.Context, attributePath *tftypes.AttributePath) (types.String, error) {
	var ret types.String
	schemaTypes := makeAttributeTypesObject(ctx, s.Schema)

	state, err := s.Raw.Unmarshal(schemaTypes)
	if err != nil {
		return ret, fmt.Errorf("error unmarshalling raw state: %s", err)
	}

	attr, remaining, err := tftypes.WalkAttributePath(state, attributePath)
	if err != nil {
		return ret, fmt.Errorf("error walking attribute path in state; %v still remains in the path: %s", remaining, err)
	}

	typ := types.StringType{}
	attrValue, err := typ.ValueFromTerraform(ctx, attr.(tftypes.Value))
	if err != nil {
		return ret, fmt.Errorf("error converting from tftypes.Value: %s", err)
	}

	ret = attrValue.(types.String)

	return ret, nil
}

// Get attempts to return the state value at attributePath. The value will be
// converted to an AttributeValue by calling the ValueFromTerraform function
// of the passed in attrType. However, the return value must still be cast to
// the desired AttributeValue type before use.
func (s *State) Get(ctx context.Context, attributePath *tftypes.AttributePath, attrType attribute.AttributeType) (attribute.AttributeValue, error) {
	schemaTypes := makeAttributeTypesObject(ctx, s.Schema)

	state, err := s.Raw.Unmarshal(schemaTypes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling raw state: %s", err)
	}

	attr, remaining, err := tftypes.WalkAttributePath(state, attributePath)
	if err != nil {
		return nil, fmt.Errorf("error walking attribute path in state; %v still remains in the path: %s", remaining, err)
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, attr.(tftypes.Value))
	if err != nil {
		return nil, fmt.Errorf("error converting from tftypes.Value: %s", err)
	}

	return attrValue, nil
}

func (s *State) Set(ctx context.Context, attributePath *tftypes.AttributePath, val interface{}) error {
	//hmm
	return nil
}
