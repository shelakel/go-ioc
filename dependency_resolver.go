package ioc

import (
	"reflect"
	"sync"
)

// RecursionLimit specifies the maximum count resolve can be called for a type and name
// before an error is raised, to avoid infinite recursion.
//
// A limit of 1 is sufficient for detecting infinite recursion
// for the Per Container and Per Scope lifetimes.
//
// An (infinite -1) limit is required for detecting infinite recursion
// for the Per Request lifetime.
//
// Due to the infinite limit required to accurately detect infinite recursion
// for the Per Request lifetime, you should set the
// RecursionLimit to a reasonable setting.
var RecursionLimit int = 30

// dependencyResolverGraph tracks the calls to resolve for a type and name
// to detect infinite recursion.
type dependencyResolverGraph struct {
	m      *sync.Mutex
	lookup map[reflect.Type]map[string]int
}

// newDependencyResolverGraph creates a new dependencyResolverGraph.
func newDependencyResolverGraph() *dependencyResolverGraph {
	return &dependencyResolverGraph{new(sync.Mutex), make(map[reflect.Type]map[string]int)}
}

// Tracks the number of times resolve is called for a type and name.
//
// Returns true while the count is less than the RecursionLimit.
func (g *dependencyResolverGraph) track(typ reflect.Type, name string) bool {
	g.m.Lock()
	if named, ok := g.lookup[typ]; ok {
		var count int
		if count, ok = named[name]; ok {
			count += 1
			if count >= RecursionLimit {
				return false
			}
		} else {
			count = 1
		}
		named[name] = count
	} else {
		g.lookup[typ] = map[string]int{name: 1}
	}
	g.m.Unlock()
	return true
}

// dependencyResolver tracks the resolve calls for a type and name, and proxies resolve calls to a Container.
//
// dependencyResolver uses a dependencyResolverGraph to track the calls to resolve for a type and name
// to detect infinite recursion.
//
// Resolve calls within a factory function are passed either the current (scoped) dependency resolver or
// a new root container level dependency resolver inheriting
// the dependencyResolverGraph from the parent dependencyResolver.
type dependencyResolver struct {
	c *Container
	g *dependencyResolverGraph
}

// newDependencyResolver creates a new newDependencyResolver.
func newDependencyResolver(c *Container, g *dependencyResolverGraph) *dependencyResolver {
	return &dependencyResolver{c, g}
}

var typeContainer = reflect.TypeOf((*Container)(nil)).Elem()
var typeFactory = reflect.TypeOf((*Factory)(nil)).Elem()

// Resolve a named instance by type with arguments.
//
// The dependencyResolver keeps track of the resolution graph to detect infinite recursion on calls to resolve for a type and name.
//
// dependencyResolver calls GetNamedSetter(v, name).
//
// Returns an error when:
//	- The value type is nil.
//	- The value isn't a pointer.
//	- The value is a nil pointer e.g. (*string)(nil) (use a pointer to a (nil) pointer instead)
//	- The dependency can't be resolved (not registered).
//	- The instance lifetime isn't supported. Currently only PerContainer, PerScope and PerRequest lifetimes are supported.
//	- An error was returned when (*Registration).CreateInstance was called.
//	- Infinite recursion is detected on a repetitive call to resolve an instance by type and name.
func (resolver *dependencyResolver) ResolveNamed(v interface{}, name string) error {
	instanceSetter, err := GetNamedSetter(v, name)
	if err != nil {
		return err
	}
	typ := instanceSetter.Type()
	if name == "" {
		switch typ {
		case typeContainer:
			instance := reflect.ValueOf(resolver.c).Elem()
			instanceSetter.Set(instance)
			return nil
		case typeFactory:
			var factory Factory = resolver
			instance := reflect.ValueOf(factory)
			instanceSetter.Set(instance)
			return nil
		}
	}
	// get the registration
	registration := resolver.c.r.get(typ, name)
	var instance *reflect.Value
	if registration == nil {
		// try to resolve using the scoped container values
		if instance = resolver.c.get(typ, name); instance != nil {
			instanceSetter.Set(*instance)
			return nil
		}
		return errUnresolvedDependency(typ, name)
	}
	switch registration.Lifetime {
	case PerContainer:
		// create a dependency resolver for the root container
		resolver1 := resolver
		if resolver.c.root != nil {
			resolver1 = newDependencyResolver(resolver.c.root, resolver.g)
		}
		// the root dependency resolver should be used to resolve
		// dependencies inside the factory function (*Registration).CreateInstance.
		// further dependency resolution will occur at the root container scope
		// i.e. no instances from the scoped container are available
		instance, err = resolver1.resolveSingletonLifetime(registration)
	case PerScope:
		instance, err = resolver.resolveSingletonLifetime(registration)
	case PerRequest:
		instance, err = resolver.resolvePerRequestLifetime(registration)
	default:
		return errUnsupportedLifetime(registration.Type, registration.Name, registration.Lifetime)
	}
	if err != nil {
		return err
	}
	instanceSetter.Set(*instance)
	return nil
}

// resolve a singleton instance for the Per Container and Per Scope lifetimes.
func (resolver *dependencyResolver) resolveSingletonLifetime(registration *Registration) (*reflect.Value, error) {
	if instance := resolver.c.instances.get(registration.Type, registration.Name); instance != nil {
		return instance, nil
	}
	if !resolver.g.track(registration.Type, registration.Name) {
		return nil, errResolveInfiniteRecursion(registration.Type, registration.Name)
	}
	instance, err := registration.CreateInstance(resolver)
	if err != nil {
		return nil, err
	}
	resolver.c.instances.set(registration.Type, registration.Name, instance)
	return instance, nil
}

// resolve an instance for the Per Request lifetime.
func (resolver *dependencyResolver) resolvePerRequestLifetime(registration *Registration) (*reflect.Value, error) {
	if !resolver.g.track(registration.Type, registration.Name) {
		return nil, errResolveInfiniteRecursion(registration.Type, registration.Name)
	}
	return registration.CreateInstance(resolver)
}
