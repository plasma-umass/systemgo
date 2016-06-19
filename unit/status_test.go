package unit

import (
	"fmt"
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

func TestStatus(t *testing.T) {

	st := Status{
		Load: LoadStatus{
			Path:   "Path",
			Loaded: Loaded,
			State:  Enabled,
			Vendor: Enabled,
		},
		Activation: ActivationStatus{
			State: Active,
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

	if st.String() != expected {
		t.Errorf(test.Mismatch, st, expected)
	}
}
