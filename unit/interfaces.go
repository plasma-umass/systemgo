package unit

// Supervisable is an interface that makes different fields of the underlying definition accesible
type Supervisable interface {
	Description() string
	Documentation() string

	Wants() []string
	Requires() []string
	Before() []string
	After() []string
	Conflicts() []string

	WantedBy() []string
	RequiredBy() []string
}

type Subber interface {
	Active() Activation
	Sub() string
}

type StartStopper interface {
	Starter
	Stopper
}
type Starter interface {
	Start() error
}
type Stopper interface {
	Stop() error
}
type Reloader interface {
	Reload() error
}
