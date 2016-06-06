package system

import (
	"strings"
	"time"

	"github.com/b1101/systemgo/lib/errors"
	"github.com/b1101/systemgo/lib/state"
	"github.com/b1101/systemgo/lib/status"
)

type System struct {
	// Map containing all units found
	Units map[string]*Unit

	// Map of booleans *Unit->bool, indicating which units are enabled
	Enabled map[*Unit]bool

	// Map of booleans *Unit->bool, indicating which units are enabled
	Failed map[*Unit]bool

	// Slice of units in the queue
	Queue *Queue

	// Status of global state
	State state.System

	// Deal with concurrency
	//sync.Mutex

	// Paths, where the unit file specifications get searched for
	Paths []string

	// Starting time
	Since time.Time
}

//var queue = make(chan *Unit)

//type units map[string]*Unit

//func (us units) String() (out string) {
//for _, u := range us {
//out += fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t\n",
//u.Name(), u.Loaded(), u.Active(), u.Sub(), u.Description())
//}
//return
//}

func New(paths ...string) (s *System, err error) {
	s = &System{}
	s.Since = time.Now()
	s.Paths = paths
	if s.Units, err = ParseAll(paths...); err != nil {
		s.State = state.Degraded
	}
	s.Queue = NewQueue()
	go s.queueStarter()
	return
}

func (s *System) Start(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		u.Log("added to the queue")
		s.Queue.Add(u)
	}
	return
}
func (s *System) Stop(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		s.Queue.Remove(u)
		u.Stop()
	}
	return
}
func (s *System) Restart(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		s.Queue.Remove(u)
		u.Stop()
		s.Queue.Add(u)
	}
	return
}
func (s *System) Reload(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		if reloader, ok := u.Supervisable.(Reloader); ok {
			reloader.Reload()
		} else {
			err = errors.NotFound
		}
	}
	return
}
func (s *System) Enable(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		s.Enabled[u] = true
	}
	return
}
func (s *System) Disable(name string) (err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		delete(s.Enabled, u)
	}
	return
}

//func (s System) WriteStatus(output io.Writer, names ...string) (err error) {
//if len(names) == 0 {
//w := tabwriter.Writer
//out += fmt.Sprintln("unit\t\t\t\tload\tactive\tsub\tdescription")
//out += fmt.Sprintln(s.Units)
//}

//var u *Unit
//for _, name := range names {
//if u, err = s.unit(name); err != nil {
//return
//}
//st := u.Status()

//st.Load.State, _ = s.IsEnabled(name)

//st.Load.Vendor.State = state.Enabled

//out += fmt.Sprintf("%s - %s\n%s\n", u.Name(), u.Description(), st)

//b := make([]byte, 1000)
//if n, _ := u.Read(b); n > 0 {
//out += fmt.Sprintf("Log:\n%s\n", b)
//}
//}

//return
//}

func (s System) Status() (st status.System) {
	return status.System{
		State:  s.State,
		Jobs:   s.Queue.Len(),
		Failed: len(s.Failed),
		Since:  s.Since,
	}
}

func (s System) StatusOf(name string) (st status.Unit, err error) {
	var u *Unit
	if u, err = s.unit(name); err != nil {
		return
	}

	var enabled state.Enable
	var vendor state.Enable

	if s.Enabled[u] {
		enabled = state.Enabled
	}

	var log []string
	b := make([]byte, 10000)

	if n, err := u.Read(b); err == nil && n > 0 {
		log = strings.Split(string(b[:n]), "\n")
	}

	return status.Unit{
		status.Load{u.Path(), u.Loaded(), enabled, status.Vendor{vendor}},
		status.Activation{u.Active(), u.Sub()},
		log,
	}, nil
}

//func (s System) WriteStatus(w io.Writer, names ...string) (n int64, err error) {
//if len(names) == 0 {
//return 0, errors.WIP
//}
//for _, name := range names {
//var s status.Unit
//if s, err = s.StatusOf(name); err != nil {
//return
//}
//_n, err = s.WriteTo(w)
//n += _n
//if err != nil {
//return
//}
//}
//return
//}

//func (s System) IsEnabled(w io.Writer, names ...string) (n int64, err error) {
//if len(names) == 0 {
//return 0, errors.WIP // TODO: Too few arguments
//}
//var st state.Enable
//if st, err = s.IsEnabled(name); err != nil {
//return
//}
//}

func (s System) IsEnabled(name string) (st state.Enable, err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil && s.Enabled[u] {
		st = state.Enabled
	}
	return
}
func (s System) IsActive(name string) (st state.Activation, err error) {
	var u *Unit
	if u, err = s.unit(name); err == nil {
		st = u.Active()
	}
	return
}

func (s System) unit(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = s.Units[name]; !ok {
		err = errors.NotFound
	}
	return
}

func isUp(u Supervisable) bool {
	return u.Active() == state.Active
}

func isLoading(u Supervisable) bool {
	return u.Active() == state.Activating
}

func (s *System) queueStarter() {
	for u := range s.Queue.Start {
		go func(u *Unit) {
			u.Log("Starting", u.Name())

			u.Log("Checking Conflicts...", u.Name())
			for _, name := range u.Conflicts() {
				if dep, _ := s.unit(name); dep != nil && isUp(dep) {
					u.Log("Unit conflicts with", name)
					return
				}
			}

			u.Log("Checking Requires...", u.Name())
			for _, name := range u.Requires() {
				if dep, err := s.unit(name); err != nil {
					u.Log(name, err.Error())
					return
				} else if !isUp(dep) && !isLoading(dep) {
					s.Queue.Add(dep)
				}
			}

			u.Log("Checking After...", u.Name())
			for _, name := range u.After() {
				u.Log("after", name)
				if dep, err := s.unit(name); err != nil {
					u.Log(name, err.Error())
					return
				} else if !isUp(dep) {
					u.Log("Waiting for", dep.Name(), "to start")
					<-dep.waitFor()
					u.Log(dep.Name(), "started")
				}
			}

			u.Log("Checking Requires again...", u.Name())
			for _, name := range u.Requires() {
				if dep, _ := s.unit(name); !isUp(dep) {
					return
				}
			}

			if err := u.Start(); err != nil {
				u.Log(err.Error())
			}

			u.Log("Started")
			u.ready()
		}(u)
	}
}
