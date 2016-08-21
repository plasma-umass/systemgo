package system

import (
	"io"

	"github.com/plasma-umass/systemgo/unit"
)

const (
	active = "active"
	dead   = "dead"
)

// Target unit type is used for grouping units
type Target struct {
	unit.Definition
	System *Daemon
}

// Define attempts to fill the targ definition by parsing r
func (targ *Target) Define(r io.Reader) (err error) {
	return unit.ParseDefinition(r, &targ.Definition)
}

// Active returns activation status of the unit
func (targ *Target) Active() unit.Activation {
	encountered := map[unit.Activation]bool{}

	for _, name := range targ.Definition.Unit.Requires {
		dep, err := targ.System.Unit(name)
		if err != nil {
			return unit.Inactive
		}
		encountered[dep.Active()] = true
	}

	for _, state := range []unit.Activation{unit.Failed, unit.Activating, unit.Deactivating, unit.Reloading, unit.Inactive} {
		if encountered[state] {
			return state
		}
	}

	return unit.Active
}

func (targ *Target) Sub() string {
	if unit.IsActive(targ) {
		return active
	}
	return dead
}
