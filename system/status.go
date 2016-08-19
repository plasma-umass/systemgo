package system

import (
	"fmt"
	"io/ioutil"
	"time"
)

// System status
type Status struct {
	// State of the system
	State State `json:"State,string"`

	// Number of queued jobs in-total
	Jobs int `json:"Jobs"`

	// Number of failed units
	Failed int `json:"Failed"`

	// Init time
	Since time.Time `json:"Since"`

	// Log
	Log []byte `json:"Log, omitempty"`
}

func (s Status) String() (out string) {
	defer func() {
		if len(s.Log) > 0 {
			out += fmt.Sprintf("\nLog:\n%s\n", s.Log)
		}
	}()
	return fmt.Sprintf(
		`State: %s
Jobs: %v queued
Failed: %v units
Since: %v`,
		s.State, s.Jobs, s.Failed, s.Since)
}

// Status returns status of the system
// If error is returned it is going to be an error,
// returned by the call to ioutil.ReadAll(sys.Log)
func (sys *Daemon) Status() (st Status, err error) {
	st = Status{Since: sys.since}

	for _, u := range sys.Units() {
		switch {
		case u.job == nil:
			continue
		case u.job.IsRunning():
			st.Jobs++
		case u.job.Failed():
			st.Failed++
		}
	}

	if st.Failed > 0 {
		st.State = Degraded
	} else {
		st.State = Running
	}

	st.Log, err = ioutil.ReadAll(sys.Log)

	return
}
