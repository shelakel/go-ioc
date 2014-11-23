package ioc

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

// ErrorCode represents an error code for distinguishing between errors.
type ErrorCode int

const (
	// ErrInstanceNotFound is raised by (*Values).GetNamed when an instance isn't found.
	ErrInstanceNotFound ErrorCode = iota
	// ErrNilType is raised by GetNamedSetter, GetNamedInstance, GetNamedType when the type of v is nil.
	// (e.g. called GetNamedType(v:nil, name:"").
	ErrNilType
	// ErrCreateInstanceNil is raised
	//   by (*Container).RegisterNamed when createInstance is nil or
	//   by (*Registration).CreateInstance when (*Registration).CreateInstanceFn is nil.
	ErrCreateInstanceNil
	// ErrCreateInstance is raised by (*Registration).CreateInstance when (*Registration).CreateInstanceFn
	// is not nil and returned an error.
	ErrCreateInstance
	// ErrUnresolvedDependency is raised by by (*dependencyResolver).ResolveNamed
	// when an instance isn't registered as a singleton or a factory function
	// and an instance can't be found on (*Container).Values.
	ErrUnresolvedDependency
	// ErrUnsupportedLifetime is raised by (*Container).RegisterNamed, (*Registration).CreateInstance
	// when the lifetime isn't supported.
	ErrUnsupportedLifetime
	// ErrUnexpectedValueType is raised by (*Registration).CreateInstance
	// when the type of the created instance doesn't match the registration type.
	ErrUnexpectedValueType
	// ErrInterfaceNotImplemented is raised by (*Registration).CreateInstance
	// when the registration type is an interface and the created instance type
	// doesn't implement the interface.
	ErrInterfaceNotImplemented
	// ErrNilValue is raised by GetNamedInstance when v is nil or a pointer to a nil value.
	ErrNilValue
	// ErrNonSetNilPointer is raised by GetNamedSetter when v is a nil pointer.
	ErrNonSetNilPointer
	// ErrRequirePointer is raised by GetNamedSetter, GetNamedType when v isn't a pointer.
	ErrRequirePointer
	// ErrResolveInfiniteRecursion is raised by (*dependencyResolver).ResolveNamed
	// when the count of resolve by type and name within a (*Container).ResolveNamed call
	// exceeds the RecursionLimit.
	ErrResolveInfiniteRecursion
)

type Error struct {
	Type      reflect.Type
	Name      string
	OtherType reflect.Type
	Code      ErrorCode
	Message   string
	Inner     error
	File      string
	LineNo    int
	Method    string
}

func (e *Error) Error() string {
	var b bytes.Buffer
	b.WriteString(e.Message)
	if e.Inner != nil {
		b.WriteRune('\n')
		b.WriteString(e.Inner.Error())
	}
	return b.String()
}

// callers: values.go
func errInstanceNotFound(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("named instance \"%s\" ", name))
	} else {
		b.WriteString("instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\" not found.", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrInstanceNotFound,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: reflect.go
func errNilType(name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("\"%s\" ", name))
	}
	b.WriteString("value type is nil.")
	return &Error{
		Name:    name,
		Code:    ErrNilType,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: container.go, registry.go
func errCreateInstanceFnNil(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: (*Registration).CreateInstanceFn is nil. unable to create ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("a named instance \"%s\" ", name))
	} else {
		b.WriteString("an instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\".", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrCreateInstanceNil,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: container.go, registry.go
func errCreateInstance(typ reflect.Type, name string, err error) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: unable to create ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("a named instance \"%s\" ", name))
	} else {
		b.WriteString("an instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\".", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrCreateInstance,
		Inner:   err,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: dependency_resolver.go
func errUnresolvedDependency(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("named instance \"%s\" ", name))
	} else {
		b.WriteString("instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\" can't be resolved.", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrUnresolvedDependency,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: container.go, dependency_resolver.go
func errUnsupportedLifetime(typ reflect.Type, name string, lifetime Lifetime) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: unsupported lifetime \"%s\". unable to create ", method, lifetime))
	if name != "" {
		b.WriteString(fmt.Sprintf("a named instance \"%s\" ", name))
	} else {
		b.WriteString("an instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\".", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrUnsupportedLifetime,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: registry.go
func errUnexpectedValueType(typ reflect.Type, name string, expectedType reflect.Type) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: expected ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("a named instance \"%s\" ", name))
	} else {
		b.WriteString("an instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\", but got \"%s\".", expectedType, typ))
	return &Error{
		Type:      typ,
		OtherType: expectedType,
		Name:      name,
		Code:      ErrUnexpectedValueType,
		Message:   b.String(),
		File:      file,
		LineNo:    lineNo,
		Method:    callingMethod,
	}
}

// callers: registry.go
func errInterfaceNotImplemented(typ reflect.Type, name string, interfaceType reflect.Type) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: expected ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("a named instance \"%s\" ", name))
	} else {
		b.WriteString("an instance ")
	}
	b.WriteString(fmt.Sprintf("implementing \"%s\", but got \"%s\".", interfaceType, typ))
	return &Error{
		Type:      typ,
		OtherType: interfaceType,
		Name:      name,
		Code:      ErrInterfaceNotImplemented,
		Message:   b.String(),
		File:      file,
		LineNo:    lineNo,
		Method:    callingMethod,
	}
}

// callers: reflect.go
func errNilValue(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("\"%s\" ", name))
	}
	b.WriteString(fmt.Sprintf("value of type \"%s\" ", typ))
	b.WriteString("is nil.")
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrNilValue,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: reflect.go
func errNonSetNilPointer(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("\"%s\" ", name))
	}
	b.WriteString(fmt.Sprintf("value of type \"%s\" ", typ))
	b.WriteString("is nil and can't be set. ")
	b.WriteString("pass a non-nil pointer or a non-nil pointer to a pointer.")
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrNonSetNilPointer,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: reflect.go
func errRequirePointer(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("\"%s\" ", name))
	}
	b.WriteString(fmt.Sprintf("value of type \"%s\" ", typ))
	b.WriteString("must be a non-nil pointer.")
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrRequirePointer,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

// callers: dependency_resolver.go
func errResolveInfiniteRecursion(typ reflect.Type, name string) error {
	method, callingMethod, file, lineNo := getCaller()
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("ioc: %s: infinite recursion detected. ", method))
	if name != "" {
		b.WriteString(fmt.Sprintf("named instance \"%s\" ", name))
	} else {
		b.WriteString("instance ")
	}
	b.WriteString(fmt.Sprintf("of type \"%s\" can't be resolved.", typ))
	return &Error{
		Type:    typ,
		Name:    name,
		Code:    ErrResolveInfiniteRecursion,
		Message: b.String(),
		File:    file,
		LineNo:  lineNo,
		Method:  callingMethod,
	}
}

//-----------------------------------------------
// helpers
//-----------------------------------------------

var pkgName = reflect.TypeOf(Values{}).PkgPath()

func getCaller() (method, callingMethod, file string, lineNo int) {
	done := false
	for i := 2; ; i++ {
		pc, f, ln, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callingMethod = runtime.FuncForPC(pc).Name()
		file = f
		lineNo = ln
		if done {
			callingMethod = path.Base(callingMethod)
			ix := strings.IndexRune(callingMethod, '.')
			callingMethod = callingMethod[ix+1:]
			break
		}
		if !strings.HasPrefix(callingMethod, pkgName) {
			done = true
			continue
		}
		if method == "" || !strings.HasSuffix(file, "_test.go") {
			method = callingMethod[len(pkgName)+1:]
		}
	}
	return
}
