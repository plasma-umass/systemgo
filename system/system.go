package system

import (
	"fmt"
	"sync"
	"time"

	"github.com/b1101/systemgo/unit"
)

type State struct {
	// Map containing all units found
	Units map[string]unit.Supervisable

	// Slice of all loaded units
	Loaded map[unit.Supervisable]bool

	// Status of global state
	State state

	// Number of failed units
	Failed failed

	// Number of queued jobs
	Jobs jobs

	// Init time
	Since time.Time

	// Deal with concurrency
	sync.Mutex
}

//func (u units) String() string {
//out := fmt.Sprint("unit\t\tload\tactive\tsub\tdescription\n")
////for _, sv := range u.Units {
////out += string(sv.)
////}
//return out // TODO: draw a fancy table
//}

type failed int

func (f failed) String() string {
	return fmt.Sprintf("%v units", int(f))
}

type jobs int

func (j jobs) String() string {
	return fmt.Sprintf("%v queued", int(j))
}

type state int

const (
	Something = iota // TODO: find all possible states
	Degraded
)
