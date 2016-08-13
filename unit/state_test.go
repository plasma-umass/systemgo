package unit_test

import (
	"fmt"
	"testing"

	"github.com/rvolosatovs/systemgo/unit"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, state.String(), out)
	}
}
