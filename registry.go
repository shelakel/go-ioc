package ioc

import (
	"fmt"
	"reflect"
	"sync"
)

// Lifetime represents the lifetime characteristics of an instance.
type Lifetime int

const (
	// Per Container Lifetime requires that an instance is only created once per container.
	PerContainer Lifetime = iota
	// Per Scope lifetime requires that an instance is only created once per scope.
	PerScope
	// Per Request lifetime requires that a new instance is created on every request.
	PerRequest
)

func (lifetime Lifetime) String() string {
	switch lifetime {
	case PerContainer:
		return "Per Container Lifetime"
	case PerScope:
		return "Per Scope Lifetime"
	case PerRequest:
		return "Per Request Lifetime"
	default:
		return fmt.Sprintf("%+v", int(lifetime))
	}
}

//-----------------------------------------------
// registry
//-----------------------------------------------

// registry is a thread safe type-name-registration container.
type registry struct {
	m             *sync.RWMutex
	registrations map[reflect.Type]map[string]*Registration
}

// newRegistry creates a new registry.
func newRegistry() *registry {
	return &registry{
		m:             new(sync.RWMutex),
		registrations: make(map[reflect.Type]map[string]*Registration),
	}
}

// Get a registration by type and name.
func (r *registry) get(typ reflect.Type, name string) *Registration {
	// assume typ != nil
	r.m.RLock()
	var registration *Registration
	if named, ok := r.registrations[typ]; ok {
		registration = named[name]
	}
	r.m.RUnlock()
	return registration
}

// Add or update a registration by type and name.
func (r *registry) set(typ reflect.Type, name string, registration *Registration) {
	// assume typ != nil
	r.m.Lock()
	if named, ok := r.registrations[typ]; ok {
		named[name] = registration
	} else {
		r.registrations[typ] = map[string]*Registration{name: registration}
	}
	r.m.Unlock()
}

// Get all the registrations.
func (r *registry) getAll() []*Registration {
	r.m.RLock()
	registrations := make([]*Registration, 0)
	for _, named := range r.registrations {
		for _, registration := range named {
			registrations = append(registrations, registration)
		}
	}
	r.m.RUnlock()
	return registrations
}

// Clone the registry.
func (r *registry) clone() *registry {
	r.m.RLock()
	clone := newRegistry()
	registrations := clone.registrations
	for k, named := range r.registrations {
		namedClone := make(map[string]*Registration, len(named))
		for name, registration := range named {
			namedClone[name] = registration
		}
		registrations[k] = namedClone
	}
	r.m.RUnlock()
	return clone
}

//-----------------------------------------------
// registration
//-----------------------------------------------

// Registration contains the information necessary to construct an instance.
type Registration struct {
	Type             reflect.Type
	Name             string
	Value            interface{}
	CreateInstanceFn func(Factory) (interface{}, error)
	Lifetime         Lifetime
}

// CreateInstance creates an instance using the factory function.
//
// Returns an error when:
//	- The factory function is nil or returns an error. (Registration.CreateInstanceFn)
//	- The created instance type is nil. (no type information)
//	- The created instance is a nil pointer or interface.
//	- The created instance type doesn't match the registration type or
//	- The implementation type is an interface and the created instance doesn't implement the interface.
func (r *Registration) CreateInstance(factory Factory) (*reflect.Value, error) {
	if r.CreateInstanceFn == nil {
		return nil, errCreateInstanceFnNil(r.Type, r.Name)
	}
	instance, err := r.CreateInstanceFn(factory)
	if err != nil {
		return nil, errCreateInstance(r.Type, r.Name, err)
	}
	rv, err := GetNamedInstance(instance, r.Name)
	if err != nil {
		return nil, err
	}
	typ := rv.Type()
	if typ == r.Type {
		return rv, nil
	}
	if r.Type == nil || r.Type.Kind() != reflect.Interface {
		return nil, errUnexpectedValueType(typ, r.Name, r.Type)
	}
	typ = reflect.TypeOf(instance)
	if !typ.Implements(r.Type) {
		return nil, errInterfaceNotImplemented(typ, r.Name, r.Type)
	}
	v := reflect.ValueOf(instance)
	return &v, nil
}
