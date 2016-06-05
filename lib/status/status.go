package status

import (
	"fmt"
	"time"

	"github.com/b1101/systemgo/lib/state"
)

type Unit struct {
	Load       Load       `json:"Load"`
	Activation Activation `json:"Activation"`
	Log        []string   `json:"Log,omitempty"`
}

type Load struct {
	Path   string       `json:"Path"`
	Loaded state.Load   `json:"Loaded"`
	State  state.Enable `json:"State"`
	Vendor Vendor       `json:"Vendor"`
}
type Activation struct {
	State state.Activation `json:"State"`
	Sub   state.Sub        `json:"Sub"`
}
type Vendor struct {
	State state.Enable `json:"State"`
}

// System status
type System struct {
	// State of the system
	State state.System `json:"State"`

	// Number of queued jobs in-total
	Jobs int `json:"Jobs"`

	// Number of failed units
	Failed int `json:"Failed"`

	// Init time
	Since time.Time `json:"Since"`

	// CGroup
	CGroup CGroup `json:"CGroup,omitempty"`
}

type CGroup interface{} // TODO: WIP

////
//// TODO: move to Systemctl
////

func (s Unit) String() (out string) {
	out = fmt.Sprintf(
		`Loaded: %s
Active: %s`,
		s.Load, s.Activation)
	if len(s.Log) > 0 {
		out += "\nLog:\n"
		for _, line := range s.Log {
			out += line + "\n"
		}
	}
	return
}

//type failed int
//
//func (f failed) String() string {
//	return fmt.Sprintf("%v units", int(f))
//}
//
//type jobs int
//
//func (j jobs) String() string {
//	return fmt.Sprintf("%v queued", int(j))
//}
//
//func (s Unit) String() string {
//	return fmt.Sprintf(`Loaded: %s
//Active: %s`, s.Load, s.Activation)
//}
//
//func (s Load) String() string {
//	return fmt.Sprintf("%s (%s; %s; %s)",
//		s.Loaded, s.Path, s.State, s.Vendor)
//}
//
//func (s Vendor) String() string {
//	return fmt.Sprintf("vendor preset: %s",
//		s.State)
//}
//
//func (s Activation) String() string {
//	return fmt.Sprintf("%s (%s)",
//		s.State, s.Sub)
//}
