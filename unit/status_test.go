package unit_test

import (
	"fmt"
	"testing"

	"github.com/b1101/systemgo/unit"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {

	st := unit.Status{
		Load: unit.LoadStatus{
			Path:   "Path",
			Loaded: unit.Loaded,
			State:  unit.Enabled,
			Vendor: unit.Enabled,
		},
		Activation: unit.ActivationStatus{
			State: unit.Active,
			Sub:   "Sub",
		},
		Log: []byte(`123456 test
654321 status`),
	}

	expected := fmt.Sprintf(
		`Loaded: %s (%s; %s; vendor preset: %s)
Active: %s (%s)
Log:
%s`,
		st.Load.Loaded, st.Load.Path, st.Load.State, st.Load.Vendor,
		st.Activation.State, st.Activation.Sub,
		st.Log,
	)

	assert.Equal(t, st.String(), expected)
}
