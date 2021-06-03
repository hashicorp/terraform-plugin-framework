package reflect

// TODO: uncomment when we have map type support
/*
func TestReflectMap_string(t *testing.T) {
	t.Parallel()

	var m map[string]string

	expected := map[string]string{
		"a": "red",
		"b": "blue",
		"c": "green",
	}

	result, err := reflectMap(context.Background(), testMapType{
		ElemType: testStringType{},
	}, tftypes.NewValue(tftypes.Map{
		AttributeType: tftypes.String,
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "red"),
		"b": tftypes.NewValue(tftypes.String, "blue"),
		"c": tftypes.NewValue(tftypes.String, "green"),
	}), reflect.ValueOf(m), Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	reflect.ValueOf(&m).Elem().Set(result)
	for k, v := range expected {
		if got, ok := m[k]; !ok {
			t.Errorf("Expected %q to be set to %q, wasn't set", k, v)
		} else if got != v {
			t.Errorf("Expected %q to be %q, got %q", k, v, got)
		}
	}
}
*/
