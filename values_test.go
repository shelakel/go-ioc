package ioc_test

import (
	. "github.com/shelakel/ioc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// to test:
// SetNamed (MustSetNamed/Set/MustSet calls SetNamed(v, name))
// GetNamed (MustGetNamed/Get/MustGet calls GetNamed(v, name))
// NewValuesScope

var _ = Describe("Values", func() {
	var values *Values
	BeforeEach(func() { values = NewValues() })
	It("should get/set simple values", func() {
		// int
		values.MustSetNamed(1, "")
		var vint int
		values.MustGetNamed(&vint, "")
		Expect(vint).To(Equal(1))
		// string
		values.MustSetNamed("test", "")
		var vstr string
		values.MustGetNamed(&vstr, "")
		Expect(vstr).To(Equal("test"))
	})
	It("should get/set struct values", func() {
		type V struct{ name string }
		values.MustSetNamed(V{name: "test"}, "")
		var v V
		values.MustGetNamed(&v, "")
		Expect(v.name).To(Equal("test"))
	})
	It("should get/set multiple anonymous struct values "+
		"to show that anonymous struct types are unique", func() {
		values.MustSetNamed(struct{ name string }{name: "test"}, "")
		values.MustSetNamed(struct{ count int }{count: 1}, "")
		var v1 struct{ name string }
		values.MustGetNamed(&v1, "")
		Expect(v1.name).To(Equal("test"))
		var v2 struct{ count int }
		values.MustGetNamed(&v2, "")
		Expect(v2.count).To(Equal(1))
	})
	It("should get/set interface values", func() {
		type IN interface{}
		var vin IN = 1
		// interfaces must be passed as a pointer,
		// otherwise the type information is lost
		// when the interface is converted to interface{}
		values.MustSetNamed(&vin, "")
		var v IN
		values.MustGetNamed(&v, "")
		Expect(v).ToNot(BeNil())
		v1, ok := v.(int)
		Expect(ok).To(BeTrue())
		Expect(v1).To(Equal(1))
	})
	It("should override the last value set", func() {
		values.MustSetNamed(1, "")
		values.MustSetNamed(2, "")
		var v int
		values.MustGetNamed(&v, "")
		Expect(v).To(Equal(2))
	})
	It("should return an error when instance not found", func() {
		var v int
		err := values.GetNamed(&v, "")
		Expect(err).ToNot(BeNil())
	})
	It("should implement Factory", func() {
		// compile time check
		var factory Factory = NewValues()
		Expect(factory).ToNot(BeNil())
	})
	It("should read value from parent", func() {
		values.MustSetNamed(1, "")      // on root values
		values = NewValuesScope(values) // check parent
		var v int
		values.MustGetNamed(&v, "")
		Expect(v).To(Equal(1))
	})
	It("should read value from ancestor", func() {
		values.MustSetNamed(1, "")      // on root values
		values = NewValuesScope(values) // check parent
		values = NewValuesScope(values) // check parent/parent
		var v int
		values.MustGetNamed(&v, "")
		Expect(v).To(Equal(1))
	})
	It("should return instance not found", func() {
		scopedValues := NewValuesScope(values)
		scopedValues.MustSetNamed(1, "")
		var v int
		err := values.GetNamed(&v, "")
		Expect(err).ToNot(BeNil())
	})
})
