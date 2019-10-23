package ioc

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// to test
// RegisterNamed (MustRegisterNamed/Register/MustRegister calls RegisterNamed(v, name))
// RegisterNamedInstance (MustRegisterNamedInstance calls RegisterNamedInstance(v, name))
// ResolveNamed: (MustResolveNamed/Resolve/MustResolve calls ResolveNamed(v, name)) + infinite recursion
// Scope()
// - resolve lifetime (per root container, per [scoped] container, per request)
// - values must be scoped
// Supported lifetimes (PerContainer, PerScope, PerRequest)

var _ = Describe("Container", func() {
	var (
		container, rootContainer *Container
	)
	BeforeEach(func() {
		container = NewContainer()
		rootContainer = container
	})

	basicSingletonTests := func(lifetime Lifetime) {
		// when RegisterInstance is used on the root or a scoped container,
		// the instance is always stored on the root container
		// see:
		// if lifetime != PerScope { return }
		// ensure the value is available from the root container,
		// even though it was registered on a scoped container.
		It("should register/resolve simple singletons", func() {
			// int
			container.MustRegisterInstance(1)
			var vint int
			container.MustResolve(&vint)
			Expect(vint).To(Equal(1))
			// string
			container.MustRegisterInstance("test")
			var vstr string
			container.MustResolve(&vstr)
			Expect(vstr).To(Equal("test"))
			if lifetime != PerScope {
				return
			}
			vint = 0
			rootContainer.MustResolve(&vint)
			Expect(vint).To(Equal(1))
			vstr = ""
			rootContainer.MustResolve(&vstr)
			Expect(vstr).To(Equal("test"))
		})
		It("should register/resolve struct singletons", func() {
			type V struct{ name string }
			container.MustRegisterInstance(V{name: "test"})
			var v V
			container.MustResolve(&v)
			Expect(v.name).To(Equal("test"))
			if lifetime != PerScope {
				return
			}
			v = V{}
			rootContainer.MustResolve(&v)
			Expect(v.name).To(Equal("test"))
		})
		It("should register/resolve multiple anonymous struct singletons "+
			"to show that anonymous struct types are unique", func() {
			container.MustRegisterInstance(struct{ name string }{name: "test"})
			container.MustRegisterInstance(struct{ count int }{count: 1})
			var v1 struct{ name string }
			container.MustResolve(&v1)
			Expect(v1.name).To(Equal("test"))
			var v2 struct{ count int }
			container.MustResolve(&v2)
			Expect(v2.count).To(Equal(1))
			if lifetime != PerScope {
				return
			}
			v1 = struct{ name string }{}
			rootContainer.MustResolve(&v1)
			Expect(v1.name).To(Equal("test"))
			v2 = struct{ count int }{}
			rootContainer.MustResolve(&v2)
			Expect(v2.count).To(Equal(1))
		})
		It("should register/resolve interface singletons", func() {
			type IN interface{}
			var vin IN = 1
			// interfaces must be passed as a pointer,
			// otherwise the type information is lost
			// when the interface is converted to interface{}
			container.MustRegisterInstance(&vin)
			var v IN
			container.MustResolve(&v)
			Expect(v).ToNot(BeNil())
			v1, ok := v.(int)
			Expect(ok).To(BeTrue())
			Expect(v1).To(Equal(1))
			if lifetime != PerScope {
				return
			}
			v = nil
			rootContainer.MustResolve(&v)
			Expect(v).ToNot(BeNil())
			v1, ok = v.(int)
			Expect(ok).To(BeTrue())
			Expect(v1).To(Equal(1))
		})
		It("should register/resolve named singletons", func() {
			container.MustRegisterNamedInstance(1, "one")
			container.MustRegisterNamedInstance(2, "two")
			var one, two int
			container.MustResolveNamed(&one, "one")
			container.MustResolveNamed(&two, "two")
			Expect(one).To(Equal(1))
			Expect(two).To(Equal(2))
			if lifetime != PerScope {
				return
			}
			one, two = 0, 0
			rootContainer.MustResolveNamed(&one, "one")
			rootContainer.MustResolveNamed(&two, "two")
			Expect(one).To(Equal(1))
			Expect(two).To(Equal(2))
		})
		It("should override the last singleton set", func() {
			container.MustRegisterInstance(1)
			container.MustRegisterInstance(2)
			var v int
			container.MustResolve(&v)
			Expect(v).To(Equal(2))
			if lifetime != PerScope {
				return
			}
			v = 0
			rootContainer.MustResolve(&v)
			Expect(v).To(Equal(2))
		})
		It("should override the last named singleton set", func() {
			container.MustRegisterNamedInstance(1, "one")
			container.MustRegisterNamedInstance(2, "one")
			var v int
			container.MustResolveNamed(&v, "one")
			Expect(v).To(Equal(2))
			if lifetime != PerScope {
				return
			}
			v = 0
			rootContainer.MustResolveNamed(&v, "one")
			Expect(v).To(Equal(2))
		})
		Context("should return an error when", func() {
			It("instance not registered", func() {
				var v int
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
		})
	}

	basicFactoryTests := func(lifetime Lifetime) {
		It("should register/resolve simple instances", func() {
			// int
			container.MustRegister(func(factory Factory) (interface{}, error) { return 1, nil }, (*int)(nil), lifetime)
			var vint int
			container.MustResolve(&vint)
			Expect(vint).To(Equal(1))
			// string
			container.MustRegister(func(factory Factory) (interface{}, error) { return "test", nil }, (*string)(nil), lifetime)
			var vstr string
			container.MustResolve(&vstr)
			Expect(vstr).To(Equal("test"))
		})
		It("should register/resolve struct instances", func() {
			type V struct{ name string }
			container.MustRegister(func(factory Factory) (interface{}, error) { return V{name: "test"}, nil }, (*V)(nil), lifetime)
			var v V
			container.MustResolve(&v)
			Expect(v.name).To(Equal("test"))
		})
		It("should register/resolve multiple anonymous struct instances "+
			"to show that anonymous struct types are unique", func() {
			container.MustRegister(func(factory Factory) (interface{}, error) { return struct{ name string }{name: "test"}, nil }, (*struct{ name string })(nil), lifetime)
			container.MustRegister(func(factory Factory) (interface{}, error) { return struct{ count int }{count: 1}, nil }, (*struct{ count int })(nil), lifetime)
			var v1 struct{ name string }
			container.MustResolve(&v1)
			Expect(v1.name).To(Equal("test"))
			var v2 struct{ count int }
			container.MustResolve(&v2)
			Expect(v2.count).To(Equal(1))
		})
		It("should register/resolve interface instances", func() {
			type IN interface{}
			// it's not necessary to return a pointer to an interface value inside factory functions,
			// if the value returned from the factory implements the interface,
			// it will be converted to the interface.
			container.MustRegister(func(factory Factory) (interface{}, error) { return 1, nil }, (*IN)(nil), lifetime)
			var v IN
			container.MustResolve(&v)
			Expect(v).ToNot(BeNil())
			v1, ok := v.(int)
			Expect(ok).To(BeTrue())
			Expect(v1).To(Equal(1))
		})
		It("should register/resolve named instances", func() {
			container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return 1, nil }, (*int)(nil), "one", lifetime)
			container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return 2, nil }, (*int)(nil), "two", lifetime)
			var one, two int
			container.MustResolveNamed(&one, "one")
			container.MustResolveNamed(&two, "two")
			Expect(one).To(Equal(1))
			Expect(two).To(Equal(2))
		})
		It("should override the last instance factory registered", func() {
			container.MustRegister(func(factory Factory) (interface{}, error) { return 1, nil }, (*int)(nil), lifetime)
			container.MustRegister(func(factory Factory) (interface{}, error) { return 2, nil }, (*int)(nil), lifetime)
			container.MustRegisterInstance(2)
			var v int
			container.MustResolve(&v)
			Expect(v).To(Equal(2))
		})
		It("should override the last named instance factory registered", func() {
			container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return 1, nil }, (*int)(nil), "one", lifetime)
			container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return 2, nil }, (*int)(nil), "one", lifetime)
			var v int
			container.MustResolveNamed(&v, "one")
			Expect(v).To(Equal(2))
		})
		Context("lifetime", func() {
			It("should return the a cached instance if not Per Request lifetime, "+
				"or an instance per container for Per Scope lifetimes and "+
				"always the same instance for Per Container lifetimes.", func() {
				x := 1
				container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return x, nil }, (*int)(nil), "", lifetime)
				var v int
				container.MustResolveNamed(&v, "")
				Expect(v).To(Equal(1))
				switch lifetime {
				case PerContainer:
					// Per Container Lifetime requires that an instance is only created once per container.
					x = 2
					container.MustResolveNamed(&v, "") // same scope
					Expect(v).To(Equal(1))
					scopedContainer := container.Scope()
					x = 3
					scopedContainer.MustResolveNamed(&v, "") // different scope
					Expect(v).To(Equal(1))
				case PerScope:
					// Per Scope lifetime requires that an instance is only created once per scope.
					x = 2
					container.MustResolveNamed(&v, "") // same scope
					Expect(v).To(Equal(1))
					x = 2
					rootContainer.MustResolveNamed(&v, "") // different scope
					Expect(v).To(Equal(2))
					x = 3
					scopedContainer := container.Scope()
					scopedContainer.MustResolveNamed(&v, "") // different scope
					Expect(v).To(Equal(3))
				case PerRequest:
					// a new instance must always be returned
					x = 2
					container.MustResolveNamed(&v, "") // same scope
					Expect(v).To(Equal(2))
					x = 3
					container.MustResolveNamed(&v, "") // different scope
					Expect(v).To(Equal(3))
				}
			})
		})
		Context("should return an error when", func() {
			It("instance factory is nil", func() {
				err := container.RegisterNamed(nil, (*string)(nil), "", lifetime)
				Expect(err).ToNot(BeNil())
			})
			It("instance factory not registered", func() {
				var v int
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
			It("instance factory returns wrong value type", func() {
				container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return "wrong", nil }, (*int)(nil), "", lifetime)
				var v int
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
			It("instance factory returns value that doesn't implement interface", func() {
				container.MustRegisterNamed(func(factory Factory) (interface{}, error) { return "wrong", nil }, (*io.Reader)(nil), "", lifetime)
				var v io.Reader
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
			It("infinite recursion is detected", func() {
				container.MustRegisterNamed(func(factory Factory) (interface{}, error) {
					var v int
					if err := Resolve(factory, &v); err != nil {
						return nil, err
					}
					return v, nil
				}, (*int)(nil), "", lifetime)
				var v int
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
			It("an error was returned when (*Registration).CreateInstance was called.", func() {
				container.MustRegisterNamed(func(factory Factory) (interface{}, error) {
					return nil, fmt.Errorf("Something went wrong")
				}, (*int)(nil), "", lifetime)
				var v int
				err := container.ResolveNamed(&v, "")
				Expect(err).ToNot(BeNil())
			})
		})
	}

	Context("per container lifetime", func() {
		Context("singleton instances", func() { basicSingletonTests(PerContainer) })
		Context("factory function instances", func() { basicFactoryTests(PerContainer) })
	})

	Context("per scope lifetime", func() {
		BeforeEach(func() {
			container = container.Scope()
		})

		Context("singleton instances", func() { basicSingletonTests(PerScope) })
		Context("factory function instances", func() { basicFactoryTests(PerScope) })
	})

	Context("per request lifetime", func() {
		Context("factory function instances", func() { basicFactoryTests(PerRequest) })
	})

	Context("should return an error when", func() {
		It("instance lifetime isn't supported", func() {
			err := container.RegisterNamed(func(factory Factory) (interface{}, error) {
				var v int
				if err := Resolve(factory, &v); err != nil {
					return nil, err
				}
				return v, nil
			}, (*int)(nil), "", Lifetime(6))
			Expect(err).ToNot(BeNil())
		})
	})
})
