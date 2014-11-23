package ioc

import "reflect"

// Get the reflect.Value that can be used to set the value of v.
//
// Returns an error when:
//	- The value type is nil. (v was passed as nil with no type information)
//	- The value isn't a pointer. (required to set v to the instance)
//	- The value is a nil pointer which can't be set. (use a pointer to a (nil) pointer instead)
func GetNamedSetter(v interface{}, name string) (*reflect.Value, error) {
	if typ := reflect.TypeOf(v); typ == nil {
		return nil, errNilType(name)
	}
	rv := reflect.ValueOf(v)
	// because reflect.TypeOf(v) can't be nil and
	// rv.Kind() must be a pointer,
	// it's not necessary to ensure rv.Kind() != reflect.Invalid
	// * reflect.Invalid = nil and zero value
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			if !rv.CanSet() {
				return nil, errNonSetNilPointer(rv.Type(), name)
			}
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if !rv.CanSet() {
		return nil, errRequirePointer(rv.Type(), name)
	}
	return &rv, nil
}

// Get the non-pointer reflect.Value of v.
//
// GetNamedInstance is used to get value and type information for storing singleton instances on the Container.
//
// Returns an error when:
//	- The value type is nil. (v was passed as nil with no type information)
//	- The value is a nil pointer or interface.
func GetNamedInstance(v interface{}, name string) (*reflect.Value, error) {
	if typ := reflect.TypeOf(v); typ == nil {
		return nil, errNilType(name)
	}
	rv := reflect.ValueOf(v)
	// because reflect.TypeOf(v) can't be nil and
	// non-nil zero values are allowed,
	// it's not necessary to ensure rv.Kind() != reflect.Invalid
	if (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) &&
		rv.IsNil() {
		return nil, errNilValue(rv.Type(), name)
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		if (rv.Kind() == reflect.Ptr ||
			rv.Kind() == reflect.Interface) &&
			rv.IsNil() {
			return nil, errNilValue(rv.Type(), name)
		}
	}
	return &rv, nil
}

// Get the non-pointer reflect.Type of v.
//
// GetNamedType is used to get the type information of implementation types.
//
// Returns an error when:
//	- The value type is nil. (v was passed as nil with no type information)
//	- The value isn't a pointer. (enforced rule to ensure Interface types are registered properly)
func GetNamedType(v interface{}, name string) (reflect.Type, error) {
	typ := reflect.TypeOf(v)
	if typ == nil {
		return nil, errNilType(name)
	}
	if typ.Kind() != reflect.Ptr {
		return nil, errRequirePointer(typ, name)
	}
	for typ.Elem().Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.Elem(), nil
}
