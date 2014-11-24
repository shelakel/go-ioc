# Go Package - ioc (beta) [![Build Status](https://travis-ci.org/shelakel/go-ioc.png?branch=master)](https://travis-ci.org/shelakel/go-ioc) [![GoDoc](http://godoc.org/github.com/shelakel/go-ioc?status.png)](http://godoc.org/github.com/shelakel/go-ioc)

Package ioc provides inversion of control containers, reflection helpers (GetNamedSetter, GetNamedInstance, GetNamedType) and Factory helpers (Resolve/ResolveNamed).

The ioc.Container and ioc.Values structs implement the Factory interface.

Package ioc is designed with the following goals in mind:

 - Well defined behavior (noted in comments and ensured by tests)
 - Idiomatic Go usage (within reason)
 - Cater for different use cases (e.g. ResolveNamed within Factory functions, MustResolvedNamed for web request scopes)
 - Robust runtime behavior (crash safe)
   - avoid infinite recursion on resolve
   - raise errors on configuration and behavioral errors
   - custom ioc.Error struct for diagnosing configuration and behavioral issues, containing the following metadata:
     calling function, file and line number, called function on package ioc, requested type and name, and an error code representing the type of error.
 - Predictable and reasonably efficient performance and memory usage (*to be ensured by benchmark tests)

_Package ioc is not dependent on the [net/http package](http://golang.org/pkg/net/http)._

Please note that package ioc is in beta, but of production quality.

Installation
------------

### Current Version: 0.1.0 Beta
### Go Version: 1.2+

```sh
go get -u github.com/shelakel/go-ioc
```
Documentation
-------------

See [GoDoc on Github](http://godoc.org/github.com/shelakel/go-ioc)

License
------------------

This project is under the MIT License. See the [LICENSE](https://github.com/shelakel/go-ioc/blob/master/LICENSE) file for the full license text.

Usage
-----

Please see [GoDoc on Github](http://godoc.org/github.com/shelakel/go-ioc) and *_test.go files.

Performance
-----------

Due to the use of the [reflect package](http://golang.org/pkg/reflect/),
ioc.Container and ioc.Values are not well suited for temporary storage (e.g. passing state to functions on a hot path).

More benchmarks to be added during optimization.

    go test -run=XXX -bench=. -benchmem=true

See [reflect_cpu_prof_latest.svg](https://github.com/shelakel/go-ioc/blob/master/reflect_cpu_prof_latest.svg) for a CPU profile of the runtime reflection functions.

| Benchmark | Iterations | Avg | Alloc | # Alloc |
| :-------- | ---------: | --: | ----: | ------: |
| BenchmarkGetNamedSetter_Int | 20000000 | 92.0 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedSetter_String | 20000000 | 92.9 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedSetter_Interface | 20000000 | 96.5 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedSetter_AnonStruct | 20000000 | 114 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedSetter_Struct | 20000000 | 93.3 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedSetter_DblPtr | 20000000 | 109 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_Int | 20000000 | 102 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_String | 20000000 | 102 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_Interface | 20000000 | 108 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_AnonStruct | 20000000 | 120 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_Struct | 20000000 | 101 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedInstance_DblPtr | 20000000 | 120 ns/op | 33 B/op | 1 allocs/op |
| BenchmarkGetNamedType_Int | 100000000 | 20.8 ns/op | 0 B/op | 0 allocs/op |
| BenchmarkGetNamedType_String | 100000000 | 20.9 ns/op | 0 B/op | 0 allocs/op |
| BenchmarkGetNamedType_Interface | 100000000 | 20.8 ns/op | 0 B/op | 0 allocs/op |
| BenchmarkGetNamedType_AnonStruct | 50000000 | 33.4 ns/op | 0 B/op | 0 allocs/op |
| BenchmarkGetNamedType_Struct | 100000000 | 20.9 ns/op | 0 B/op | 0 allocs/op |
| BenchmarkGetNamedType_DblPtr | 50000000 | 33.4 ns/op | 0 B/op | 0 allocs/op |

Preliminary benchmarking on my machine (i7-4770K, 2400MHz RAM) yielded 200 ns per *Set*/*Get* operation on ioc.Values, 400 ns per cached/singleton *Resolve* and 1300 ns per request to resolve via the factory function.

Tests
-----

Tests are written using [Ginkgo](http://onsi.github.io/ginkgo/) with the [Gomega](http://onsi.github.io/gomega/) matchers.

Benchmark functions are written using the [testing](http://golang.org/pkg/testing/) package.

TODO
----

 - Improve performance.
 - Populate function that uses a Factory to populate struct instances via dependency injection on tagged fields.
   The current thinking is to support dynamic population e.g. standard ioc="constant", dynamic ioc_route="id" via ResolveNamed((*Factory),"route") ->
   (dynamic Factory).ResolveNamed((type), "id").
 - Improve README with Usage examples and topics.

Contributing
-------------

Contributions are welcome, especially in the area of additional tests and performance enhancements.

 1. Find an issue that bugs you / open a new one.
 2. Discuss.
 3. Branch off the develop branch, commit, test.
 4. Submit a pull request / attach the commits to the issue.
