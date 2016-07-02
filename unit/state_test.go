package unit_test

import (
	"fmt"
	"testing"

	"github.com/b1101/systemgo/test"
	"github.com/b1101/systemgo/unit"
)

func TestStates(t *testing.T) {
	var activation unit.Activation
	var load unit.Load
	var enable unit.Enable

	states := map[fmt.Stringer]string{
		unit.Loaded: "Loaded",
		unit.Active: "Active",
		unit.Static: "Static",

		activation: "Inactive",
		load:       "Stub",
		enable:     "Disabled",
	}

	for state, out := range states {
		if state.String() != out {
			t.Errorf(test.Mismatch, state, out)
		}
	}
}
