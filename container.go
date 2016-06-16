package ioc

// Container is an inversion of control container.
type Container struct {
	root *Container
	*Values
	r         *registry
	instances *Values
}

//-----------------------------------------------
// ctor
//-----------------------------------------------

// NewContainer creates a new inversion of control container.
func NewContainer() *Container {
	return &Container{
		Values:    NewValues(),
		r:         newRegistry(),
		instances: NewValues(),
	}
}

// Scope creates a new scoped container from the current container.
//
// The Values of the current container are scoped and the registry inherited by the scoped container.
//
// Scoped Values will resolve an instance from an ancestor when the current container is unable to resolve the instance by type and name.
func (c *Container) Scope() *Container {
	root := c
	if c.root != nil {
		root = c.root
	}
	return &Container{
		root:      root,
		Values:    NewValuesScope(c.Values),
		r:         c.r.clone(),
		instances: NewValues(),
	}
}

//-----------------------------------------------
// registry implementation
//-----------------------------------------------

// Returns the registrations for the container.
func (c *Container) Registrations() []*Registration {
	return c.r.getAll()
}

// Register an instance factory with a specific lifetime.
//
// Register calls RegisterNamed(createInstance, implType, "", lifetime).
func (c *Container) Register(createInstance func(Factory) (interface{}, error), implType interface{}, lifetime Lifetime) error {
	return c.RegisterNamed(createInstance, implType, "", lifetime)
}

// Register an instance factory with a specific lifetime.
//
// MustRegister calls Register(createInstance, implType, lifetime) and panics if an error is returned.
func (c *Container) MustRegister(createInstance func(Factory) (interface{}, error), implType interface{}, lifetime Lifetime) {
	if err := c.Register(createInstance, implType, lifetime); err != nil {
		panic(err)
	}
}

// Register a named instance factory with a specific lifetime.
//
// Returns an error when:
//	- The factory function is nil. (createInstance)
//	- The implementing type is nil.
//	- The implementing type isn't a pointer.
//	- The instance lifetime isn't supported. Currently only PerContainer, PerScope and PerRequest lifetimes are supported.
func (c *Container) RegisterNamed(createInstance func(Factory) (interface{}, error), implType interface{}, name string, lifetime Lifetime) error {
	typ, err := GetNamedType(implType, name)
	if err != nil {
		return err
	}
	if createInstance == nil {
		return errCreateInstanceFnNil(typ, name)
	}
	registration := &Registration{
		Type:             typ,
		Name:             name,
		CreateInstanceFn: createInstance,
		Lifetime:         lifetime,
	}
	// must keep the Lifetime check in sync with dependencyResolver.ResolveNamed
	if lifetime != PerContainer && lifetime != PerScope && lifetime != PerRequest {
		return errUnsupportedLifetime(registration.Type, registration.Name, lifetime)
	}
	c.r.set(typ, name, registration)
	return nil
}

// Register a named instance factory with a specific lifetime.
//
// MustRegisterNamed calls RegisterNamed(createInstance, implType, name, lifetime) and panics if an error is returned.
func (c *Container) MustRegisterNamed(createInstance func(Factory) (interface{}, error), implType interface{}, name string, lifetime Lifetime) {
	if err := c.RegisterNamed(createInstance, implType, name, lifetime); err != nil {
		panic(err)
	}
}

// Register an instance on the root container.
//
// RegisterInstance calls RegisterNamedInstance(v, "").
func (c *Container) RegisterInstance(v interface{}) error {
	return c.RegisterNamedInstance(v, "")
}

// Register an instance on the root container.
//
// MustRegisterInstance calls RegisterInstance(v) and panics if an error is returned.
func (c *Container) MustRegisterInstance(v interface{}) {
	if err := c.RegisterInstance(v); err != nil {
		panic(err)
	}
}

// Register a named instance on the root container.
//
// Returns an error when:
//	- The instance type is nil.
//	- The instance is a nil pointer or interface.
func (c *Container) RegisterNamedInstance(v interface{}, name string) error {
	instance, err := GetNamedInstance(v, name)
	if err != nil {
		return err
	}
	typ := instance.Type()
	createInstance := func(Factory) (interface{}, error) {
		return v, nil
	}
	registration := &Registration{
		Type:             typ,
		Name:             name,
		Value:            v,
		CreateInstanceFn: createInstance,
		Lifetime:         PerContainer,
	}
	c.r.set(typ, name, registration)
	root := c.root
	if root == nil {
		root = c
	}
	root.instances.set(typ, name, instance)
	return nil
}

// Register a named instance on the root container.
//
// MustRegisterNamedInstance calls RegisterNamedInstance(v, name) and panics if an error is returned.
func (c *Container) MustRegisterNamedInstance(v interface{}, name string) {
	if err := c.RegisterNamedInstance(v, name); err != nil {
		panic(err)
	}
}

//-----------------------------------------------
// factory implementation
//-----------------------------------------------

// Resolve an instance by type.
//
// Resolve calls c.ResolveNamed(v, "").
func (c *Container) Resolve(v interface{}) error {
	return c.ResolveNamed(v, "")
}

// Resolve an instance by type.
//
// MustResolve calls Resolve(v) and panics if an error is returned.
func (c *Container) MustResolve(v interface{}) {
	if err := c.Resolve(v); err != nil {
		panic(err)
	}
}

// Resolve a named instance by type.
//
// ResolveNamed creates a dependency resolver implementing the Factory interface, that proxies resolve calls to the Container.
//
// The dependency resolver is passed to instance factory functions (instead of the container) and keeps track
// of the resolve call history for the request to detect infinite recursion.
//
// Returns an error when:
//	- The value type is nil.
//	- The value isn't a pointer.
//	- The value is a nil pointer e.g. (*string)(nil) (use a pointer to a (nil) pointer instead)
//	- The dependency can't be resolved (not registered).
//	- The instance lifetime isn't supported. Currently only PerContainer, PerScope and PerRequest lifetimes are supported.
//	- An error was returned when (*Registration).CreateInstance was called.
//	- Infinite recursion is detected on a repetitive call to resolve an instance by type and name.
func (c *Container) ResolveNamed(v interface{}, name string) error {
	resolver := newDependencyResolver(c, newDependencyResolverGraph())
	return resolver.ResolveNamed(v, name)
}

// Resolve a named instance by type.
//
// MustResolveNamed calls ResolveNamed and panics if an error is returned.
func (c *Container) MustResolveNamed(v interface{}, name string) {
	if err := c.ResolveNamed(v, name); err != nil {
		panic(err)
	}
}
