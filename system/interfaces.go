package system

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
