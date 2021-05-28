package objects

import "reflect"

// IsNilValue performs a panic safe nil check on v, including embedded interface values
func IsNilValue(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer:
		return rv.IsNil()
	}

	return false
}

// IsUnmarshalableValue performs a panic safe basic type and value assertion on v
// to check if it is apt for Unmarshal operations (that is, a non nil pointer value)
func IsUnmarshalableValue(v interface{}) bool {
	if v == nil {
		return false
	}

	entityType := reflect.TypeOf(v)
	entityKind := entityType.Kind()
	if entityKind != reflect.Ptr || IsNilValue(v) {
		return false
	}

	return true
}
