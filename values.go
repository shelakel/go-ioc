package ioc

import (
	"reflect"
	"sync"
)

// Values is a thread safe type-name-instance container.
type Values struct {
	parent    *Values
	m         *sync.RWMutex
	instances map[reflect.Type]map[string]*reflect.Value
}

//-----------------------------------------------
// ctor
//-----------------------------------------------

// NewValues creates a new Values struct.
func NewValues() *Values {
	return NewValuesScope(nil)
}

// NewValuesScope creates a new scoped Values struct with a reference to a parent Values struct.
//
// Get calls will check the ancestors to resolve the instance by type and name.
func NewValuesScope(parent *Values) *Values {
	return &Values{parent, new(sync.RWMutex), make(map[reflect.Type]map[string]*reflect.Value)}
}

//-----------------------------------------------
// private methods
//-----------------------------------------------

// Get an instance by type and name.
//
// Returns nil if the instance wasn't found.
func (values *Values) get(typ reflect.Type, name string) *reflect.Value {
	// assume typ != nil
	values.m.RLock()
	var instance *reflect.Value
	if named, ok := values.instances[typ]; ok {
		instance = named[name]
	}
	values.m.RUnlock()
	return instance
}

// Get an instance by type and name recursively from the parent Values struct.
//
// Returns nil if the instance can't be found on an ancestor.
func (values *Values) getParent(typ reflect.Type, name string) *reflect.Value {
	// assume typ != nil
	parent := values.parent
	var instance *reflect.Value
	for parent != nil {
		if instance = parent.get(typ, name); instance != nil {
			break
		}
		parent = parent.parent
	}
	return instance
}

// Add or update an instance by type and name.
func (values *Values) set(typ reflect.Type, name string, instance *reflect.Value) {
	// assume typ != nil and instance != nil
	values.m.Lock()
	if named, ok := values.instances[typ]; ok {
		named[name] = instance
	} else {
		values.instances[typ] = map[string]*reflect.Value{name: instance}
	}
	values.m.Unlock()
}

//-----------------------------------------------
// public methods
//-----------------------------------------------

// Get an instance by type.
//
// Get calls GetNamed(v, "").
func (values *Values) Get(v interface{}) error {
	return values.GetNamed(v, "")
}

// Get an instance by type.
//
// MustGet calls Get(v) and panics if an error is returned.
func (values *Values) MustGet(v interface{}) {
	if err := values.Get(v); err != nil {
		panic(err)
	}
}

// Get a named instance by type.
//
// GetNamed calls GetNamedSetter(v, name).
//
// Returns an error when:
//	- The value type is nil. (v was passed as nil with no type information)
//	- The value isn't a pointer. (required to set v to the instance)
//	- The value is a nil pointer which can't be set. (use a pointer to a (nil) pointer instead)
//	- The instance isn't found.
func (values *Values) GetNamed(v interface{}, name string) error {
	instanceSetter, err := GetNamedSetter(v, name)
	if err != nil {
		return err
	}
	typ := instanceSetter.Type()
	instance := values.get(typ, name)
	if instance == nil {
		if instance = values.getParent(typ, name); instance == nil {
			return errInstanceNotFound(typ, name)
		}
	}
	instanceSetter.Set(*instance)
	return nil
}

// Get a named instance by type.
//
// MustGetNamed calls GetNamed(v, name) and panics if an error is returned.
func (values *Values) MustGetNamed(v interface{}, name string) {
	if err := values.GetNamed(v, name); err != nil {
		panic(err)
	}
}

// Set an instance by type.
//
// Set calls SetNamed(v, "").
func (values *Values) Set(v interface{}) error {
	return values.SetNamed(v, "")
}

// Set an instance by type.
//
// MustSet calls Set(v) and panics if an error is returned.
func (values *Values) MustSet(v interface{}) {
	if err := values.Set(v); err != nil {
		panic(err)
	}
}

// Set a named instance by type.
//
// SetNamed calls GetNamedInstance(v, name).
//
// Returns an error when:
//	- The value type is nil. (v was passed as nil with no type information)
//	- The value is nil.		 (interface or pointer)
func (values *Values) SetNamed(v interface{}, name string) error {
	instance, err := GetNamedInstance(v, name)
	if err != nil {
		return err
	}
	typ := instance.Type()
	values.set(typ, name, instance)
	return nil
}

// Set a named instance by type.
//
// MustSetNamed calls SetNamed(v, name) and panics if an error is returned.
func (values *Values) MustSetNamed(v interface{}, name string) {
	if err := values.SetNamed(v, name); err != nil {
		panic(err)
	}
}

//-----------------------------------------------
// factory implementation
//-----------------------------------------------

// Resolve a named instance by type.
//
// ResolveNamed calls GetNamed(v, name).
func (values *Values) ResolveNamed(v interface{}, name string) error {
	return values.GetNamed(v, name)
}
