package ioc

// Resolve uses a factory to resolve instances by type.
func Resolve(factory Factory, instances ...interface{}) error {
	for _, instance := range instances {
		if err := factory.ResolveNamed(instance, ""); err != nil {
			return err
		}
	}
	return nil
}

// MustResolve uses a factory to resolve instances by type.
//
// MustResolve calls Resolve(factory, instances...) and panics if an error is returned.
func MustResolve(factory Factory, instances ...interface{}) {
	if err := Resolve(factory, instances...); err != nil {
		panic(err)
	}
}

// ResolveNamed uses a factory to resolve instances by type and name.
func ResolveNamed(factory Factory, namedInstances map[string][]interface{}) error {
	for name, instances := range namedInstances {
		for _, instance := range instances {
			if err := factory.ResolveNamed(instance, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// MustResolveNamed uses a factory to resolve instances by type and name.
//
// MustResolveNamed calls ResolveNamed(factory, namedInstances) and panics if an error is returned.
func MustResolveNamed(factory Factory, namedInstances map[string][]interface{}) {
	if err := ResolveNamed(factory, namedInstances); err != nil {
		panic(err)
	}
}
