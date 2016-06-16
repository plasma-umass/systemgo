package unit

import (
	"fmt"
	"testing"
)

func TestStates(t *testing.T) {
	var activation Activation
	var load Load
	var enable Enable

	states := map[fmt.Stringer]string{
		Loaded: "Loaded",
		Active: "Active",
		Static: "Static",

		activation: "Inactive",
		load:       "Stub",
		enable:     "Disabled",
	}

	for state, out := range states {
		if state.String() != out {
			t.Errorf("%s != %s", state, out)
		}
	}
}
