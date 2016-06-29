package system

//go:generate mockgen -source=interfaces.go -package=system -destination=mock_interfaces_test.go

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
