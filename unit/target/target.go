package target

import (
	"io"

	"github.com/b1101/systemgo/unit"
)

// Target unit type.
// Is different enough from other units to not include
// it in the unit package
type Unit struct {
	// Target definition does not have any specific fields
	unit.Definition

	// Used to get target dependencies
	Get GetFunc
}

type GetFunc func(name string) (unit.Subber, error)

func NewTarget(fn GetFunc) (targ *Unit) {
	return &Unit{
		Get: fn,
	}
}

// Define attempts to fill the targ definition by parsing r
func (targ *Unit) Define(r io.Reader) (err error) {
	return unit.ParseDefinition(r, &targ.Definition)
}

// Start attempts to start the dependencies of the target
func (targ *Unit) Start() (err error) {
	return
}

// Start attempts to stop units started by the target
func (targ *Unit) Stop() (err error) {
	return
}

func (targ *Unit) Active() unit.Activation {
	encountered := map[unit.Activation]bool{}

	for _, name := range targ.Definition.Unit.Requires {
		dep, err := targ.Get(name)
		if err != nil {
			return unit.Inactive
		}
		encountered[dep.Active()] = true
	}

	for _, state := range []unit.Activation{unit.Failed, unit.Activating, unit.Deactivating} {
		if encountered[state] {
			return state
		}
	}

	return unit.Active
}

func (targ *Unit) Sub() string {
	return "TODO"
}
