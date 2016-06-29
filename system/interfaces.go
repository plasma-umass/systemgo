package system

import "github.com/b1101/systemgo/unit"

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

type Daemon interface {
	Start(...string) error
	Stop(string) error
	Restart(string) error
	Reload(string) error
	Enable(string) error
	Disable(string) error

	Status() (Status, error)
	StatusOf(string) (unit.Status, error)
	IsEnabled(string) (unit.Enable, error)
	IsActive(string) (unit.Activation, error)

	SetPaths(...string)
}

type Manager interface {
	Get(string) (*Unit, error)
	NewJob(JobType, ...string) (*Job, error)
}

type Loader interface {
	Load(string) (*Unit, error)
}
type Getter interface {
	Get(string) (*Unit, error)
}
