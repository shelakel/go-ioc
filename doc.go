/*
Package ioc provides inversion of control containers and functionality.

The ioc.Container and ioc.Values structs implement the Factory interface.

Basics

The containers provided by package ioc make use of runtime reflection to detect value types and to resolve an instance by type and name.

Containers can be scoped.

Example:
	type UserRepository interface {
		GetById(int64 id) (*User, error)
	}

	type PostgresUserRepository struct {
		db *sql.DB
	}

	func newPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
		return &PostgresUserRepository{db: db}
	}

	func (repo *PostgresUserRepository) GetById(int64 id) (*User, error) {
		return nil, nil // stub
	}

	c := ioc.NewContainer()
	// register PostgresUserRepository as UserRepository
	createInstance := func(factory ioc.Factory) (interface{}, error) {
		var db *sql.DB
		// Resolve requires a non-nil pointer, but you can pass a reference to a nil pointer.
		if err := ioc.Resolve(factory, &db); err != nil {
			return nil, err
		}
		repo := newPostgresUserRepository(db)
		return repo, nil
	}
	implType := (*UserRepository)(nil) // must be a nil pointer
	lifetime := ioc.PerContainer
	c.MustRegister(createInstance, implType, lifetime)
	// register the singleton *sql.DB
	driverName := "postgres"
	dataSourceName := "mydb"
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	c.MustRegisterInstance(db) // the instance being registered can't be a nil pointer or interface.
	// tell the container to construct a UserRepository
	var userRepository UserRepository
	// Resolve requires a non-nil pointer, but you can pass a reference to a nil pointer.
	c.MustResolve(&userRepository)

	// scoped example
	func ContainerMiddleware(next http.Handler) http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			scopedContainer := c.Scope()
			scopedContainer.MustSet(&w)
			scopedContainer.MustSet(r)
			// use the scopedContainer within the request scope, e.g. using gorilla
			context.Set(r, "container", scopedContainer)
		})
	}

	func DoSomething(w http.ResponseWriter, r *http.Request) {
		container := context.Get(r, "container").(*Container)
		var userRepository UserRepository
		container.MustGet(&userRepository)
		user, err := userRepository.GetById(1)
	}

Resolving Instances

The following methods can be used to resolve instances:
	- (*ioc.Values) Get/GetNamed
	- (ioc.Factory) ResolveNamed
	- ioc.Resolve/ioc.ResolveNamed

Resolved instances are stored in the value pointed to by v.

	A non-nil pointer or a reference to a nil-pointer is required to set the value pointed to by v.

Registering Instances

The following methods can be used to register instances:
	- (*ioc.Values) Set/SetNamed
	- (*ioc.Container) Set/SetNamed (scoped container singleton/scope vars)
	- (*ioc.Container) RegisterInstance/RegisterNamedInstance (root container singleton)

	The instance being registered can't be a nil pointer or interface.

As a best practice, always pass the pointer to the instance you want to register.

Example: Register an instance of an interface type
	values := NewValues()

	file, _ := os.Open("my_file")
	// file: *os.File
	defer file.Close()
	var f io.Reader = file

	values.Set(f)  // type registered: *os.File (wrong!)
	values.Set(&f) // type registered: io.Reader


Instance Factory Registrations

The following methods can be used to register an instance factory:
	- (*ioc.Container) Register/RegisterNamed (instance factory)

An instance factory function must return a non-nil value or an error.

	The value type should match the implementing type or
	the implementing type must be an interface and the value must implement that interface.

The Lifetime characteristics determines how a resolved instance is reused:
	- Per Container Lifetime requires that an instance is only created once per container.
	- Per Scope lifetime requires that an instance is only created once per scope.
	- Per Request lifetime requires that a new instance is created on every request.

*/
package ioc
