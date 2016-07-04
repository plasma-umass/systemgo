package system

import (
	"fmt"

	"github.com/b1101/systemgo/unit"
)

// Supervisable represents anything that can be supervised by an instance of a system.Interface
type Supervisable interface {
	unit.StartStopper
	unit.Subber
	unit.Definer

	//Statuser

	//definition
}

// Statuser is implemented by anything that can report its' own status
type Statuser interface {
	Status() fmt.Stringer
}

// Daemon represents the system as seen from 'outer scope'.
// Manages instances of unit.Interface and operates on them by name.
// Provides handlers for system control and is meant to be exposed
// to the 'init' package, or anything else that could want to use it
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

// Interface represents the system as seen from the 'inner scope'.
// Manages instances of unit.Interface and operates on them directly.
type Interface interface {
	Manager
	Getter
}

// TODO: comment.
type Manager interface {
	NewJob(JobType, ...string) (*Job, error)
}

type Getter interface {
	Get(string) (*Unit, error)
}
type Loader interface {
	Load(string) (*Unit, error)
}

// Parser is implemented by any value that has a Parse menthod, which
// takes an io.Reader and returns a unit.Interface
type Parser interface {
	Parse(io.Reader) (unit.Interface, error)
type definition interface {
	Description() string
	Documentation() string

	Wants() []string
	Requires() []string

	After() []string
	Before() []string

	Conflicts() []string

	RequiredBy() []string
	WantedBy() []string
}
