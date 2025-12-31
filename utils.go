package dbkit

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// ToJson converts any model, struct or Record to a JSON string.
// It is designed to be highly robust and safe:
// 1. Handles nil and typed nil pointers gracefully (returns "{}").
// 2. Disables HTML escaping for cleaner output in non-browser contexts.
// 3. Catches marshaling errors to prevent panics.
func ToJson(v interface{}) string {
	if isNil(v) {
		return "{}"
	}

	// Use a buffer and encoder to customize behavior
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // Keep symbols like <, >, & as is

	if err := enc.Encode(v); err != nil {
		return "{}"
	}

	// json.Encoder adds a newline at the end, let's trim it
	res := buf.Bytes()
	if len(res) > 0 && res[len(res)-1] == '\n' {
		res = res[:len(res)-1]
	}

	return string(res)
}

// isNil checks if an interface is truly nil, including typed nil pointers.
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}
