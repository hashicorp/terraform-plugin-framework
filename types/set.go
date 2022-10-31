package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ attr.Type              = SetType{}
	_ xattr.TypeWithValidate = SetType{}
	_ attr.Value             = &Set{}
)

// SetType is an AttributeType representing a set of values. All values must
// be of the same type, which the provider must specify as the ElemType
// property.
type SetType struct {
	ElemType attr.Type
}

// ElementType returns the attr.Type elements will be created from.
func (st SetType) ElementType() attr.Type {
	return st.ElemType
}

// WithElementType returns a SetType that is identical to `l`, but with the
// element type set to `typ`.
func (st SetType) WithElementType(typ attr.Type) attr.TypeWithElementType {
	return SetType{ElemType: typ}
}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (st SetType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Set{
		ElementType: st.ElemType.TerraformType(ctx),
	}
}

// ValueFromTerraform returns an attr.Value given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (st SetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	set := Set{
		ElemType: st.ElemType,
		state:    valueStateDeprecated,
	}
	if in.Type() == nil {
		set.Null = true
		return set, nil
	}
	if !in.Type().Equal(st.TerraformType(ctx)) {
		return nil, fmt.Errorf("can't use %s as value of Set with ElementType %T, can only use %s values", in.String(), st.ElemType, st.ElemType.TerraformType(ctx).String())
	}
	if !in.IsKnown() {
		set.Unknown = true
		return set, nil
	}
	if in.IsNull() {
		set.Null = true
		return set, nil
	}
	val := []tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	elems := make([]attr.Value, 0, len(val))
	for _, elem := range val {
		av, err := st.ElemType.ValueFromTerraform(ctx, elem)
		if err != nil {
			return nil, err
		}
		elems = append(elems, av)
	}
	set.Elems = elems
	return set, nil
}

// Equal returns true if `o` is also a SetType and has the same ElemType.
func (st SetType) Equal(o attr.Type) bool {
	if st.ElemType == nil {
		return false
	}
	other, ok := o.(SetType)
	if !ok {
		return false
	}
	return st.ElemType.Equal(other.ElemType)
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// set.
func (st SetType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if _, ok := step.(tftypes.ElementKeyValue); !ok {
		return nil, fmt.Errorf("cannot apply step %T to SetType", step)
	}

	return st.ElemType, nil
}

// String returns a human-friendly description of the SetType.
func (st SetType) String() string {
	return "types.SetType[" + st.ElemType.String() + "]"
}

// Validate implements type validation. This type requires all elements to be
// unique.
func (st SetType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Is(tftypes.Set{}) {
		err := fmt.Errorf("expected Set value, received %T with value: %v", in, in)
		diags.AddAttributeError(
			path,
			"Set Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var elems []tftypes.Value

	if err := in.As(&elems); err != nil {
		diags.AddAttributeError(
			path,
			"Set Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	validatableType, isValidatable := st.ElemType.(xattr.TypeWithValidate)

	// Attempting to use map[tftypes.Value]struct{} for duplicate detection yields:
	//   panic: runtime error: hash of unhashable type tftypes.primitive
	// Instead, use for loops.
	for indexOuter, elemOuter := range elems {
		// Only evaluate fully known values for duplicates and validation.
		if !elemOuter.IsFullyKnown() {
			continue
		}

		// Validate the element first
		if isValidatable {
			elemValue, err := st.ElemType.ValueFromTerraform(ctx, elemOuter)
			if err != nil {
				diags.AddAttributeError(
					path,
					"Set Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
				)
				return diags
			}
			diags = append(diags, validatableType.Validate(ctx, elemOuter, path.AtSetValue(elemValue))...)
		}

		// Then check for duplicates
		for indexInner := indexOuter + 1; indexInner < len(elems); indexInner++ {
			elemInner := elems[indexInner]

			if !elemInner.Equal(elemOuter) {
				continue
			}

			// TODO: Point at element attr.Value when Validate method is converted to attr.Value
			// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/172
			diags.AddAttributeError(
				path,
				"Duplicate Set Element",
				fmt.Sprintf("This attribute contains duplicate values of: %s", elemInner),
			)
		}
	}

	return diags
}

// ValueType returns the Value type.
func (t SetType) ValueType(_ context.Context) attr.Value {
	return Set{
		ElemType: t.ElemType,
	}
}

// SetNull creates a Set with a null value. Determine whether the value is
// null via the Set type IsNull method.
//
// Setting the deprecated Set type ElemType, Elems, Null, or Unknown fields
// after creating a Set with this function has no effect.
func SetNull(elementType attr.Type) Set {
	return Set{
		elementType: elementType,
		state:       valueStateNull,
	}
}

// SetUnknown creates a Set with an unknown value. Determine whether the
// value is unknown via the Set type IsUnknown method.
//
// Setting the deprecated Set type ElemType, Elems, Null, or Unknown fields
// after creating a Set with this function has no effect.
func SetUnknown(elementType attr.Type) Set {
	return Set{
		elementType: elementType,
		state:       valueStateUnknown,
	}
}

// SetValue creates a Set with a known value. Access the value via the Set
// type Elements or ElementsAs methods.
//
// Setting the deprecated Set type ElemType, Elems, Null, or Unknown fields
// after creating a Set with this function has no effect.
func SetValue(elementType attr.Type, elements []attr.Value) (Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for idx, element := range elements {
		if !elementType.Equal(element.Type(ctx)) {
			diags.AddError(
				"Invalid Set Element Type",
				"While creating a Set value, an invalid element was detected. "+
					"A Set must use the single, given element type. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Set Element Type: %s\n", elementType.String())+
					fmt.Sprintf("Set Index (%d) Element Type: %s", idx, element.Type(ctx)),
			)
		}
	}

	if diags.HasError() {
		return SetUnknown(elementType), diags
	}

	return Set{
		elementType: elementType,
		elements:    elements,
		state:       valueStateKnown,
	}, nil
}

// SetValueFrom creates a Set with a known value, using reflection rules.
// The elements must be a slice which can convert into the given element type.
// Access the value via the Set type Elements or ElementsAs methods.
func SetValueFrom(ctx context.Context, elementType attr.Type, elements any) (Set, diag.Diagnostics) {
	attrValue, diags := reflect.FromValue(
		ctx,
		SetType{ElemType: elementType},
		elements,
		path.Empty(),
	)

	if diags.HasError() {
		return SetUnknown(elementType), diags
	}

	set, ok := attrValue.(Set)

	// This should not happen, but ensure there is an error if it does.
	if !ok {
		diags.AddError(
			"Unable to Convert Set Value",
			"An unexpected result occurred when creating a Set using SetValueFrom. "+
				"This is an issue with terraform-plugin-framework and should be reported to the provider developers.",
		)
	}

	return set, diags
}

// SetValueMust creates a Set with a known value, converting any diagnostics
// into a panic at runtime. Access the value via the Set
// type Elements or ElementsAs methods.
//
// This creation function is only recommended to create Set values which will
// not potentially effect practitioners, such as testing, or exhaustively
// tested provider logic.
//
// Setting the deprecated Set type ElemType, Elems, Null, or Unknown fields
// after creating a Set with this function has no effect.
func SetValueMust(elementType attr.Type, elements []attr.Value) Set {
	set, diags := SetValue(elementType, elements)

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

		panic("SetValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return set
}

// Set represents a set of attr.Value, all of the same type,
// indicated by ElemType.
type Set struct {
	// Unknown will be set to true if the entire set is an unknown value.
	// If only some of the elements in the set are unknown, their known or
	// unknown status will be represented however that attr.Value
	// surfaces that information. The Set's Unknown property only tracks
	// if the number of elements in a Set is known, not whether the
	// elements that are in the set are known.
	//
	// If the Set was created with the SetValue, SetNull, or SetUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the SetUnknown function to create an unknown Set
	// value or use the IsUnknown method to determine whether the Set value
	// is unknown instead.
	Unknown bool

	// Null will be set to true if the set is null, either because it was
	// omitted from the configuration, state, or plan, or because it was
	// explicitly set to null.
	//
	// If the Set was created with the SetValue, SetNull, or SetUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the SetNull function to create a null Set value or
	// use the IsNull method to determine whether the Set value is null
	// instead.
	Null bool

	// Elems are the elements in the set.
	//
	// If the Set was created with the SetValue, SetNull, or SetUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the SetValue function to create a known Set value or
	// use the Elements or ElementsAs methods to retrieve the Set elements
	// instead.
	Elems []attr.Value

	// ElemType is the tftypes.Type of the elements in the set. All
	// elements in the set must be of this type.
	//
	// Deprecated: Use the SetValue, SetNull, or SetUnknown functions
	// to create a Set or use the ElementType method to retrieve the
	// Set element type instead.
	ElemType attr.Type

	// elements is the collection of known values in the Set.
	elements []attr.Value

	// elementType is the type of the elements in the Set.
	elementType attr.Type

	// state represents whether the Set is null, unknown, or known. During the
	// exported field deprecation period, this state can also be "deprecated",
	// which remains the zero-value for compatibility to ensure exported field
	// updates take effect. The zero-value will be changed to null in a future
	// version.
	state valueState
}

func (s Set) ToFrameworkValue() attr.Value {
	return s
}

// Elements returns the collection of elements for the Set. Returns nil if the
// Set is null or unknown.
func (s Set) Elements() []attr.Value {
	if s.state == valueStateDeprecated {
		return s.Elems
	}

	return s.elements
}

// ElementsAs populates `target` with the elements of the Set, throwing an
// error if the elements cannot be stored in `target`.
func (s Set) ElementsAs(ctx context.Context, target interface{}, allowUnhandled bool) diag.Diagnostics {
	// we need a tftypes.Value for this Set to be able to use it with our
	// reflection code
	val, err := s.ToTerraformValue(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Set Element Conversion Error",
				"An unexpected error was encountered trying to convert set elements. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			),
		}
	}
	return reflect.Into(ctx, s.Type(ctx), val, target, reflect.Options{
		UnhandledNullAsEmpty:    allowUnhandled,
		UnhandledUnknownAsEmpty: allowUnhandled,
	}, path.Empty())
}

// ElementType returns the element type for the Set.
func (s Set) ElementType(_ context.Context) attr.Type {
	if s.state == valueStateDeprecated {
		return s.ElemType
	}

	return s.elementType
}

// Type returns a SetType with the same element type as `s`.
func (s Set) Type(ctx context.Context) attr.Type {
	return SetType{ElemType: s.ElementType(ctx)}
}

// ToTerraformValue returns the data contained in the Set as a tftypes.Value.
func (s Set) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if s.state == valueStateDeprecated && s.ElemType == nil {
		return tftypes.Value{}, fmt.Errorf("cannot convert Set to tftypes.Value if ElemType field is not set")
	}
	setType := tftypes.Set{ElementType: s.ElementType(ctx).TerraformType(ctx)}

	switch s.state {
	case valueStateDeprecated:
		if s.Unknown {
			return tftypes.NewValue(setType, tftypes.UnknownValue), nil
		}
		if s.Null {
			return tftypes.NewValue(setType, nil), nil
		}
		vals := make([]tftypes.Value, 0, len(s.Elems))
		for _, elem := range s.Elems {
			val, err := elem.ToTerraformValue(ctx)
			if err != nil {
				return tftypes.NewValue(setType, tftypes.UnknownValue), err
			}
			vals = append(vals, val)
		}
		if err := tftypes.ValidateValue(setType, vals); err != nil {
			return tftypes.NewValue(setType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(setType, vals), nil
	case valueStateKnown:
		vals := make([]tftypes.Value, 0, len(s.elements))

		for _, elem := range s.elements {
			val, err := elem.ToTerraformValue(ctx)

			if err != nil {
				return tftypes.NewValue(setType, tftypes.UnknownValue), err
			}

			vals = append(vals, val)
		}

		if err := tftypes.ValidateValue(setType, vals); err != nil {
			return tftypes.NewValue(setType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(setType, vals), nil
	case valueStateNull:
		return tftypes.NewValue(setType, nil), nil
	case valueStateUnknown:
		return tftypes.NewValue(setType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Set state in ToTerraformValue: %s", s.state))
	}
}

// Equal returns true if the Set is considered semantically equal
// (same type and same value) to the attr.Value passed as an argument.
func (s Set) Equal(o attr.Value) bool {
	other, ok := o.(Set)
	if !ok {
		return false
	}
	if s.state != other.state {
		return false
	}
	if s.state == valueStateKnown {
		if !s.elementType.Equal(other.elementType) {
			return false
		}

		if len(s.elements) != len(other.elements) {
			return false
		}

		for _, elem := range s.elements {
			if !other.contains(elem) {
				return false
			}
		}

		return true
	}
	if s.Unknown != other.Unknown {
		return false
	}
	if s.Null != other.Null {
		return false
	}
	if s.ElemType == nil && other.ElemType != nil {
		return false
	}
	if s.ElemType != nil && !s.ElemType.Equal(other.ElemType) {
		return false
	}
	if len(s.Elems) != len(other.Elems) {
		return false
	}
	for _, elem := range s.Elems {
		if !other.contains(elem) {
			return false
		}
	}
	return true
}

func (s Set) contains(v attr.Value) bool {
	for _, elem := range s.Elements() {
		if elem.Equal(v) {
			return true
		}
	}

	return false
}

// IsNull returns true if the Set represents a null value.
func (s Set) IsNull() bool {
	if s.state == valueStateNull {
		return true
	}

	return s.state == valueStateDeprecated && s.Null
}

// IsUnknown returns true if the Set represents a currently unknown value.
// Returns false if the Set has a known number of elements, even if all are
// unknown values.
func (s Set) IsUnknown() bool {
	if s.state == valueStateUnknown {
		return true
	}

	return s.state == valueStateDeprecated && s.Unknown
}

// String returns a human-readable representation of the Set value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (s Set) String() string {
	if s.IsUnknown() {
		return attr.UnknownValueString
	}

	if s.IsNull() {
		return attr.NullValueString
	}

	var res strings.Builder

	res.WriteString("[")
	for i, e := range s.Elements() {
		if i != 0 {
			res.WriteString(",")
		}
		res.WriteString(e.String())
	}
	res.WriteString("]")

	return res.String()
}
