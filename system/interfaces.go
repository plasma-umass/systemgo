package system

import (
	"fmt"

	"github.com/b1101/systemgo/unit"
)

// Supervisable represents anything that can be supervised by an instance of a system.Interface
type Supervisable interface {
	unit.Interface
	Statuser
}

// Statuser is implemented by anything that can report its' own status
type Statuser interface {
	Status() fmt.Stringer
}

// Daemon represents the system as seen from 'outer scope'.
// Manages instances of unit.Interface and operates on them by name.
// Provides handlers for system control and is meant to be exposed
// to the 'init' package, or anything else that could want to use it
type Getter interface {
	Get(string) (*Unit, error)
}
type Loader interface {
	Load(string) (*Unit, error)
}
