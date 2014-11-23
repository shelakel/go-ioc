package ioc_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/shelakel/ioc"
)

var _ = Describe("GetNamedSetter", func() {
	Context("should return an error when", func() {
		//	- The value type is nil. (v was passed as nil with no type information)
		//	- The value isn't a pointer. (required to set v to the instance)
		//	- The value is a nil pointer which can't be set. (use a pointer to a (nil) pointer instead)
		mustGetError := func(v interface{}) {
			instanceSetter, err := GetNamedSetter(v, "")
			Expect(instanceSetter).To(BeNil())
			Expect(err).ToNot(BeNil())
		}
		It("v is nil", func() { mustGetError(nil) })
		It("v isn't a pointer", func() { mustGetError(1) })
		It("v is a nil pointer which can't be set", func() { mustGetError((*string)(nil)) })
	})
	Context("should return a reflect.Value that can be set when", func() {
		mustGetSetter := func(v interface{}) {
			instanceSetter, err := GetNamedSetter(v, "")
			Expect(instanceSetter).ToNot(BeNil())
			Expect(err).To(BeNil())
		}
		It("v is a non-nil pointer", func() { i := new(int); mustGetSetter(&i) })
		It("v is a pointer to a nil pointer", func() { var i *int; mustGetSetter(&i) })
	})
})

var _ = Describe("GetNamedInstance", func() {
	Context("should return an error when", func() {
		//	- The value type is nil. (v was passed as nil with no type information)
		//	- The value is a nil pointer or interface.
		//	- The value is invalid.
		mustGetError := func(v interface{}) {
			instance, err := GetNamedInstance(v, "")
			Expect(instance).To(BeNil())
			Expect(err).ToNot(BeNil())
		}
		It("v is nil", func() { mustGetError(nil) })
		It("v is a nil pointer", func() { mustGetError((*string)(nil)) })
		It("v is a nil interface", func() { mustGetError((interface{})(nil)) })
	})
	Context("should return a non-pointer reflect.Value", func() {
		mustGetInstance := func(v interface{}) {
			instance, err := GetNamedInstance(v, "")
			Expect(instance).ToNot(BeNil())
			Expect(err).To(BeNil())
		}
		It("v is a non-nil pointer", func() { i := new(int); mustGetInstance(&i) })
		It("v is a non-nil interface", func() { type IN interface{}; var in IN = 1; mustGetInstance(&in) })
		It("v is a zero value", func() { mustGetInstance(reflect.Zero(reflect.TypeOf(1))) })
		It("v is a struct value", func() { mustGetInstance(1) })
	})
})

var _ = Describe("GetNamedType", func() {
	Context("should return an error when", func() {
		//	- The value type is nil. (v was passed as nil with no type information)
		//	- The value isn't a pointer. (enforced rule to ensure Interface types are registered properly)
		mustGetError := func(v interface{}) {
			typ, err := GetNamedType(v, "")
			Expect(typ).To(BeNil())
			Expect(err).ToNot(BeNil())
		}
		It("v is nil", func() { mustGetError(nil) })
		It("v isn't a pointer (struct)", func() { mustGetError(1) })
		// if an non-pointer interface is passed, the interface type information is lost due
		// the value getting converted to interface{}.
		// hence not possible to test that
		// an error is returned for an interface type unless a pointer to an interface is passed;
		// which would be valid.
	})
	Context("should return a non-pointer reflect.Type", func() {
		mustGetType := func(v interface{}) {
			typ, err := GetNamedType(v, "")
			Expect(typ).ToNot(BeNil())
			Expect(err).To(BeNil())
		}
		It("v is a non-nil pointer", func() { i := new(int); mustGetType(i) })
		It("v is a non-nil pointer", func() { i := new(int); mustGetType(&i) })
		It("v is a nil pointer", func() { mustGetType((*int)(nil)) })
		It("v is a non-nil interface", func() { type IN interface{}; var in IN = 1; mustGetType(&in) })
		It("v is a nil interface", func() { type IN interface{}; mustGetType((*IN)(nil)) })
	})
})

type testStruct struct{ name string }

var brv *reflect.Value
var btyp reflect.Type
var berr error

func benchGetNamedSetter(v interface{}, b *testing.B) {
	var instanceSetter *reflect.Value
	var err error
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		instanceSetter, err = GetNamedSetter(v, "")
	}
	b.StopTimer()
	brv = instanceSetter
	berr = err
}

func BenchmarkGetNamedSetter_Int(b *testing.B) {
	v := 1
	benchGetNamedSetter(&v, b)
}

func BenchmarkGetNamedSetter_String(b *testing.B) {
	v := "test"
	benchGetNamedSetter(&v, b)
}

func BenchmarkGetNamedSetter_Interface(b *testing.B) {
	type IN interface{}
	var v IN = 1
	benchGetNamedSetter(&v, b)
}

func BenchmarkGetNamedSetter_AnonStruct(b *testing.B) {
	v := &struct{ name string }{name: "test"}
	benchGetNamedSetter(&v, b)
}

func BenchmarkGetNamedSetter_Struct(b *testing.B) {
	v := testStruct{name: "test"}
	benchGetNamedSetter(&v, b)
}

func BenchmarkGetNamedSetter_DblPtr(b *testing.B) {
	x := 1
	v := &x
	benchGetNamedSetter(&v, b)
}

//-----------------------------------------------

func benchGetNamedInstance(v interface{}, b *testing.B) {
	var instance *reflect.Value
	var err error
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		instance, err = GetNamedInstance(v, "")
	}
	b.StopTimer()
	brv = instance
	berr = err
}

func BenchmarkGetNamedInstance_Int(b *testing.B) {
	v := 1
	benchGetNamedInstance(&v, b)
}

func BenchmarkGetNamedInstance_String(b *testing.B) {
	v := "test"
	benchGetNamedInstance(&v, b)
}

func BenchmarkGetNamedInstance_Interface(b *testing.B) {
	type IN interface{}
	var v IN = 1
	benchGetNamedInstance(&v, b)
}

func BenchmarkGetNamedInstance_AnonStruct(b *testing.B) {
	v := &struct{ name string }{name: "test"}
	benchGetNamedInstance(&v, b)
}

func BenchmarkGetNamedInstance_Struct(b *testing.B) {
	v := testStruct{name: "test"}
	benchGetNamedInstance(&v, b)
}

func BenchmarkGetNamedInstance_DblPtr(b *testing.B) {
	x := 1
	v := &x
	benchGetNamedInstance(&v, b)
}

//-----------------------------------------------

func benchGetNamedType(v interface{}, b *testing.B) {
	var typ reflect.Type
	var err error
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		typ, err = GetNamedType(v, "")
	}
	b.StopTimer()
	btyp = typ
	berr = err
}

func BenchmarkGetNamedType_Int(b *testing.B) {
	v := 1
	benchGetNamedType(&v, b)
}

func BenchmarkGetNamedType_String(b *testing.B) {
	v := "test"
	benchGetNamedType(&v, b)
}

func BenchmarkGetNamedType_Interface(b *testing.B) {
	type IN interface{}
	var v IN = 1
	benchGetNamedType(&v, b)
}

func BenchmarkGetNamedType_AnonStruct(b *testing.B) {
	v := &struct{ name string }{name: "test"}
	benchGetNamedType(&v, b)
}

func BenchmarkGetNamedType_Struct(b *testing.B) {
	v := testStruct{name: "test"}
	benchGetNamedType(&v, b)
}

func BenchmarkGetNamedType_DblPtr(b *testing.B) {
	x := 1
	v := &x
	benchGetNamedType(&v, b)
}
