package ioc

// Factory represents a container able to
// resolve instances by type and name.
//
// Implemented by:
//	- Values
//	- Container
//	- dependencyResolver (internal)
type Factory interface {
	// Resolve a named instance by type.
	ResolveNamed(v interface{}, name string) error
}
