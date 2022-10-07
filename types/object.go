package types

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Type  = ObjectType{}
	_ attr.Value = &Object{}
)

// ObjectType is an AttributeType representing an object.
type ObjectType struct {
	AttrTypes map[string]attr.Type
}

// WithAttributeTypes returns a new copy of the type with its attribute types
// set.
func (o ObjectType) WithAttributeTypes(typs map[string]attr.Type) attr.TypeWithAttributeTypes {
	return ObjectType{
		AttrTypes: typs,
	}
}

// AttributeTypes returns the type's attribute types.
func (o ObjectType) AttributeTypes() map[string]attr.Type {
	return o.AttrTypes
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (o ObjectType) TerraformType(ctx context.Context) tftypes.Type {
	attributeTypes := map[string]tftypes.Type{}
	for k, v := range o.AttrTypes {
		attributeTypes[k] = v.TerraformType(ctx)
	}
	return tftypes.Object{
		AttributeTypes: attributeTypes,
	}
}

// ValueFromTerraform returns an attr.Value given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (o ObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	object := Object{
		AttrTypes: o.AttrTypes,
	}
	if in.Type() == nil {
		object.Null = true
		return object, nil
	}
	if !in.Type().Equal(o.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", o.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		object.Unknown = true
		return object, nil
	}
	if in.IsNull() {
		object.Null = true
		return object, nil
	}
	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := object.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}
	object.Attrs = attributes
	return object, nil
}

// Equal returns true if `candidate` is also an ObjectType and has the same
// AttributeTypes.
func (o ObjectType) Equal(candidate attr.Type) bool {
	other, ok := candidate.(ObjectType)
	if !ok {
		return false
	}
	if len(other.AttrTypes) != len(o.AttrTypes) {
		return false
	}
	for k, v := range o.AttrTypes {
		attr, ok := other.AttrTypes[k]
		if !ok {
			return false
		}
		if !v.Equal(attr) {
			return false
		}
	}
	return true
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// object.
func (o ObjectType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.AttributeName); !ok {
		return nil, fmt.Errorf("cannot apply step %T to ObjectType", step)
	}

	return o.AttrTypes[string(step.(tftypes.AttributeName))], nil
}

// String returns a human-friendly description of the ObjectType.
func (o ObjectType) String() string {
	var res strings.Builder
	res.WriteString("types.ObjectType[")
	keys := make([]string, 0, len(o.AttrTypes))
	for k := range o.AttrTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for pos, key := range keys {
		if pos != 0 {
			res.WriteString(", ")
		}
		res.WriteString(`"` + key + `":`)
		res.WriteString(o.AttrTypes[key].String())
	}
	res.WriteString("]")
	return res.String()
}

// ValueType returns the Value type.
func (t ObjectType) ValueType(_ context.Context) attr.Value {
	return Object{
		AttrTypes: t.AttrTypes,
	}
}

// ObjectNull creates a Object with a null value. Determine whether the value is
// null via the Object type IsNull method.
//
// Setting the deprecated Object type AttrTypes, Attrs, Null, or Unknown fields
// after creating a Object with this function has no effect.
func ObjectNull(attributeTypes map[string]attr.Type) Object {
	return Object{
		attributeTypes: attributeTypes,
		state:          valueStateNull,
	}
}

// ObjectUnknown creates a Object with an unknown value. Determine whether the
// value is unknown via the Object type IsUnknown method.
//
// Setting the deprecated Object type AttrTypes, Attrs, Null, or Unknown fields
// after creating a Object with this function has no effect.
func ObjectUnknown(attributeTypes map[string]attr.Type) Object {
	return Object{
		attributeTypes: attributeTypes,
		state:          valueStateUnknown,
	}
}

// ObjectValue creates a Object with a known value. Access the value via the Object
// type ElementsAs method.
//
// Setting the deprecated Object type AttrTypes, Attrs, Null, or Unknown fields
// after creating a Object with this function has no effect.
func ObjectValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing Object Attribute Value",
				"While creating a Object value, a missing attribute value was detected. "+
					"A Object must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Object Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid Object Attribute Type",
				"While creating a Object value, an invalid attribute value was detected. "+
					"A Object must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Object Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("Object Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra Object Attribute Value",
				"While creating a Object value, an extra attribute value was detected. "+
					"A Object must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra Object Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return ObjectUnknown(attributeTypes), diags
	}

	return Object{
		attributeTypes: attributeTypes,
		attributes:     attributes,
		state:          valueStateKnown,
	}, nil
}

// ObjectValueFrom creates a Object with a known value, using reflection rules.
// The attributes must be a map of string attribute names to attribute values
// which can convert into the given attribute type or a struct with tfsdk field
// tags. Access the value via the Object type Elements or ElementsAs methods.
func ObjectValueFrom(ctx context.Context, attributeTypes map[string]attr.Type, attributes any) (Object, diag.Diagnostics) {
	attrValue, diags := reflect.FromValue(
		ctx,
		ObjectType{AttrTypes: attributeTypes},
		attributes,
		path.Empty(),
	)

	if diags.HasError() {
		return ObjectUnknown(attributeTypes), diags
	}

	m, ok := attrValue.(Object)

	// This should not happen, but ensure there is an error if it does.
	if !ok {
		diags.AddError(
			"Unable to Convert Object Value",
			"An unexpected result occurred when creating a Object using ObjectValueFrom. "+
				"This is an issue with terraform-plugin-framework and should be reported to the provider developers.",
		)
	}

	return m, diags
}

// ObjectValueMust creates a Object with a known value, converting any diagnostics
// into a panic at runtime. Access the value via the Object
// type Elements or ElementsAs methods.
//
// This creation function is only recommended to create Object values which will
// not potentially effect practitioners, such as testing, or exhaustively
// tested provider logic.
//
// Objectting the deprecated Object type ElemType, Elems, Null, or Unknown fields
// after creating a Object with this function has no effect.
func ObjectValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) Object {
	object, diags := ObjectValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("ObjectValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

var (
	_ attr.ValueWithAttrs = Object{}
)

// Object represents an object
type Object struct {
	// Unknown will be set to true if the entire object is an unknown value.
	// If only some of the elements in the object are unknown, their known or
	// unknown status will be represented however that attr.Value
	// surfaces that information. The Object's Unknown property only tracks
	// if the number of elements in a Object is known, not whether the
	// elements that are in the object are known.
	//
	// If the Object was created with the ObjectValue, ObjectNull, or ObjectUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the ObjectNull function to create a null Object value or
	// use the IsNull method to determine whether the Object value is null
	// instead.
	Unknown bool

	// Null will be set to true if the object is null, either because it was
	// omitted from the configuration, state, or plan, or because it was
	// explicitly set to null.
	//
	// If the Object was created with the ObjectValue, ObjectNull, or ObjectUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the ObjectNull function to create a null Object value or
	// use the IsNull method to determine whether the Object value is null
	// instead.
	Null bool

	// Attrs is the mapping of known attribute values in the Object.
	//
	// If the Object was created with the ObjectValue, ObjectNull, or ObjectUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the ObjectValue function to create a known Object value or
	// use the As or Attributes methods to retrieve the Object attributes
	// instead.
	Attrs map[string]attr.Value

	// AttrTypes is the mapping of attribute types in the Object. Required
	// for a valid Object.
	//
	// Deprecated: Use the ObjectValue, ObjectNull, or ObjectUnknown functions
	// to create a Object or use the AttributeTypes method to retrieve the
	// Object attribute types instead.
	AttrTypes map[string]attr.Type

	// attributes is the mapping of known attribute values in the Object.
	attributes map[string]attr.Value

	// attributeTypes is the type of the attributes in the Object.
	attributeTypes map[string]attr.Type

	// state represents whether the Object is null, unknown, or known. During the
	// exported field deprecation period, this state can also be "deprecated",
	// which remains the zero-value for compatibility to ensure exported field
	// updates take effect. The zero-value will be changed to null in a future
	// version.
	state valueState
}

func (o Object) GetAttrs() map[string]attr.Value {
	return o.Attrs
}

func (o Object) SetAttrs(attrs map[string]attr.Value) attr.ValueWithAttrs {
	o.Attrs = attrs

	return o
}

// ObjectAsOptions is a collection of toggles to control the behavior of
// Object.As.
type ObjectAsOptions struct {
	// UnhandledNullAsEmpty controls what happens when As needs to put a
	// null value in a type that has no way to preserve that distinction.
	// When set to true, the type's empty value will be used.  When set to
	// false, an error will be returned.
	UnhandledNullAsEmpty bool

	// UnhandledUnknownAsEmpty controls what happens when As needs to put
	// an unknown value in a type that has no way to preserve that
	// distinction. When set to true, the type's empty value will be used.
	// When set to false, an error will be returned.
	UnhandledUnknownAsEmpty bool
}

// As populates `target` with the data in the Object, throwing an error if the
// data cannot be stored in `target`.
func (o Object) As(ctx context.Context, target interface{}, opts ObjectAsOptions) diag.Diagnostics {
	// we need a tftypes.Value for this Object to be able to use it with
	// our reflection code
	obj := ObjectType{AttrTypes: o.AttrTypes}
	val, err := o.ToTerraformValue(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Object Conversion Error",
				"An unexpected error was encountered trying to convert object. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			),
		}
	}
	return reflect.Into(ctx, obj, val, target, reflect.Options{
		UnhandledNullAsEmpty:    opts.UnhandledNullAsEmpty,
		UnhandledUnknownAsEmpty: opts.UnhandledUnknownAsEmpty,
	}, path.Empty())
}

// Attributes returns the mapping of known attribute values for the Object.
// Returns nil if the Object is null or unknown.
func (o Object) Attributes() map[string]attr.Value {
	if o.state == valueStateDeprecated {
		return o.Attrs
	}

	return o.attributes
}

// AttributeTypes returns the mapping of attribute types for the Object.
func (o Object) AttributeTypes(_ context.Context) map[string]attr.Type {
	if o.state == valueStateDeprecated {
		return o.AttrTypes
	}

	return o.attributeTypes
}

// Type returns an ObjectType with the same attribute types as `o`.
func (o Object) Type(ctx context.Context) attr.Type {
	return ObjectType{AttrTypes: o.AttributeTypes(ctx)}
}

// ToTerraformValue returns the data contained in the attr.Value as
// a tftypes.Value.
func (o Object) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if o.state == valueStateDeprecated && o.AttrTypes == nil {
		return tftypes.Value{}, fmt.Errorf("cannot convert Object to tftypes.Value if AttrTypes field is not set")
	}
	attrTypes := map[string]tftypes.Type{}
	for attr, typ := range o.AttributeTypes(ctx) {
		attrTypes[attr] = typ.TerraformType(ctx)
	}
	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch o.state {
	case valueStateDeprecated:
		if o.Unknown {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
		}
		if o.Null {
			return tftypes.NewValue(objectType, nil), nil
		}
		vals := map[string]tftypes.Value{}

		for k, v := range o.Attrs {
			val, err := v.ToTerraformValue(ctx)
			if err != nil {
				return tftypes.NewValue(objectType, tftypes.UnknownValue), err
			}
			vals[k] = val
		}
		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(objectType, vals), nil
	case valueStateKnown:
		vals := make(map[string]tftypes.Value, len(o.attributes))

		for name, v := range o.attributes {
			val, err := v.ToTerraformValue(ctx)

			if err != nil {
				return tftypes.NewValue(objectType, tftypes.UnknownValue), err
			}

			vals[name] = val
		}

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case valueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case valueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", o.state))
	}
}

// Equal returns true if the Object is considered semantically equal
// (same type and same value) to the attr.Value passed as an argument.
func (o Object) Equal(c attr.Value) bool {
	other, ok := c.(Object)
	if !ok {
		return false
	}
	if o.state != other.state {
		return false
	}
	if o.state == valueStateKnown {
		if len(o.attributeTypes) != len(other.attributeTypes) {
			return false
		}

		for name, oAttributeType := range o.attributeTypes {
			otherAttributeType, ok := other.attributeTypes[name]

			if !ok {
				return false
			}

			if !oAttributeType.Equal(otherAttributeType) {
				return false
			}
		}

		if len(o.attributes) != len(other.attributes) {
			return false
		}

		for name, oAttribute := range o.attributes {
			otherAttribute, ok := other.attributes[name]

			if !ok {
				return false
			}

			if !oAttribute.Equal(otherAttribute) {
				return false
			}
		}

		return true
	}
	if o.Unknown != other.Unknown {
		return false
	}
	if o.Null != other.Null {
		return false
	}
	if len(o.AttrTypes) != len(other.AttrTypes) {
		return false
	}
	for k, v := range o.AttrTypes {
		attr, ok := other.AttrTypes[k]
		if !ok {
			return false
		}
		if !v.Equal(attr) {
			return false
		}
	}
	if len(o.Attrs) != len(other.Attrs) {
		return false
	}
	for k, v := range o.Attrs {
		attr, ok := other.Attrs[k]
		if !ok {
			return false
		}
		if !v.Equal(attr) {
			return false
		}
	}

	return true
}

// IsNull returns true if the Object represents a null value.
func (o Object) IsNull() bool {
	if o.state == valueStateNull {
		return true
	}

	return o.state == valueStateDeprecated && o.Null
}

// IsUnknown returns true if the Object represents a currently unknown value.
func (o Object) IsUnknown() bool {
	if o.state == valueStateUnknown {
		return true
	}

	return o.state == valueStateDeprecated && o.Unknown
}

// String returns a human-readable representation of the Object value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (o Object) String() string {
	if o.IsUnknown() {
		return attr.UnknownValueString
	}

	if o.IsNull() {
		return attr.NullValueString
	}

	// We want the output to be consistent, so we sort the output by key
	keys := make([]string, 0, len(o.Attributes()))
	for k := range o.Attributes() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var res strings.Builder

	res.WriteString("{")
	for i, k := range keys {
		if i != 0 {
			res.WriteString(",")
		}
		res.WriteString(fmt.Sprintf(`"%s":%s`, k, o.Attributes()[k].String()))
	}
	res.WriteString("}")

	return res.String()
}
