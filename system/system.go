package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/b1101/systemgo/unit"
)

var DEFAULT_PATHS = []string{"/etc/systemd/system/", "/run/systemd/system", "/lib/systemd/system"}

type System struct {
	// Map containing pointers to all active units(name -> *Unit)
	units map[string]*Unit

	// Map containing pointers to all units including those failed to load(name -> *Unit)
	loaded map[string]*Unit

	// Paths, where the unit file specifications get searched for
	paths []string

	// System state
	state State

	// Starting time
	since time.Time

	// System log
	Log *Log
}

//func (us units) String() (out string) {
//for _, u := range us {
//out += fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t\n",
//u.Name(), u.Loaded(), u.Active(), u.Sub(), u.Description())
//}
//return
//}
func New() (sys *System) {
	return &System{
		since: time.Now(),
		//queue:  NewQueue(),
		Log:   NewLog(),
		paths: DEFAULT_PATHS,
		units: make(map[string]*Unit),
	}
}

func (sys *System) SetPaths(paths ...string) {
	sys.paths = paths
}

//func (sys System) WriteStatus(output io.Writer, names ...string) (err error) {
//if len(names) == 0 {
//w := tabwriter.Writer
//out += fmt.Sprintln("unit\t\t\t\tload\tactive\tsub\tdescription")
//out += fmt.Sprintln(s.Units)
//}

func (sys *System) Unit(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.units[name]; !ok {
		err = ErrNotFound
	}
	return
}

func (sys *System) Get(name string) (u *Unit, err error) {
	if u, err = sys.Unit(name); err == ErrNotFound {
		u, err = sys.Load(name)
	}
	return
}

func (sys *System) Loaded(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.units[name]; !ok {
		err = ErrNotFound
	}
	return
}

func (sys *System) Start(names ...string) (err error) {
	var job *Job
	if job, err = sys.NewJob(start, names...); err != nil {
		return
	}

	return job.Start()
}

func (sys *System) Stop(name string) (err error) {
	var job *Job
	if job, err = sys.NewJob(stop, name); err != nil {
		return
	}

	return job.Start()
}

func (sys *System) Restart(name string) (err error) {
	// rewrite as a loop to preserve errors
	if err = sys.Stop(name); err != nil {
		return
	}
	return sys.Start(name)
}

func (sys *System) Reload(name string) (err error) {
	var u *Unit
	if u, err = sys.Unit(name); err != nil {
		return
	}

	if reloader, ok := u.Supervisable.(unit.Reloader); ok {
		return reloader.Reload()
	}

	return ErrNoReload
}

func (sys *System) Enable(name string) (err error) {
	var u *Unit
	if u, err = sys.Unit(name); err != nil {
		return
	}
	u.Log.Println("enable")
	return fmt.Errorf("TODO")
}

func (sys *System) Disable(name string) (err error) {
	var u *Unit
	if u, err = sys.Unit(name); err != nil {
		return
	}
	u.Log.Println("disable")
	return fmt.Errorf("TODO")
}

// IsEnabled returns enable state of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *System) IsEnabled(name string) (st unit.Enable, err error) {
	//var u *Unit
	//if u, err = sys.Unit(name); err == nil && sys.Enabled[u] {
	//st = unit.Enabled
	//}
	return unit.Enabled, ErrNotImplemented
}

// IsActive returns activation state of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *System) IsActive(name string) (st unit.Activation, err error) {
	var u *Unit
	if u, err = sys.Unit(name); err == nil {
		st = u.Active()
	}
	return
}

// StatusOf returns status of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *System) StatusOf(name string) (st unit.Status, err error) {
	var u *Unit
	if u, err = sys.Unit(name); err != nil {
		return
	}

	st = unit.Status{
		Load: unit.LoadStatus{
			Path:   u.Path(),
			Loaded: u.Loaded(),
			State:  unit.Enabled,
		},
		Activation: unit.ActivationStatus{
			State: u.Active(),
			Sub:   u.Sub(),
		},
	}

	st.Log, err = ioutil.ReadAll(u.Log)

	return
}
func (sys System) IsEnabled(name string) (st unit.Enable, err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil && sys.Enabled[u] {
		st = unit.Enabled
	}
	return
}
func (sys System) IsActive(name string) (st unit.Activation, err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		st = u.Active()
	}
	return
}

func (sys *System) Unit(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.loaded[name]; !ok {
		u, err = sys.load(name)
	}
	return
}

//func (sys *System) Units(names ...string) (units []*Unit, err error) {
//units = make([]*Unit, len(names))
//for i, name := range names {
//if units[i], err = sys.Unit(name); err != nil {
//return
//}
//}
//return
//}

func (sys *System) queueStarter() {
	for u := range sys.Queue.Start {
		go func(u *Unit) {
			u.Log.Println("Starting", u.Name())

			u.Log.Println("Checking Conflicts...", u.Name())
			for _, name := range u.Conflicts() {
				if dep, _ := sys.Unit(name); dep != nil && isActive(dep) {
					u.Log.Println("Unit conflicts with", name)
					return
				}
			}

			u.Log.Println("Checking Requires...", u.Name())
			for _, name := range u.Requires() {
				if dep, err := sys.Unit(name); err != nil {
					u.Log.Println(name, err.Error())
					return
				} else if !isActive(dep) && !isActivating(dep) {
					sys.Queue.Add(dep)
				}
			}

			u.Log.Println("Checking After...", u.Name())
			for _, name := range u.After() {
				u.Log.Println("after", name)
				if dep, err := sys.Unit(name); err != nil {
					u.Log.Println(name, err.Error())
					return
				} else if !isActive(dep) {
					u.Log.Println("Waiting for", dep.Name(), "to start")
					<-dep.waitFor()
					u.Log.Println(dep.Name(), "started")
				}
			}

			u.Log.Println("Checking Requires again...", u.Name())
			for _, name := range u.Requires() {
				if dep, _ := sys.Unit(name); !isActive(dep) {
					return
				}
			}

			if err := u.Start(); err != nil {
				u.Log.Println(err.Error())
			}

			u.Log.Println("Started")
			u.ready()
		}(u)
	}
}

func isActive(u Supervisable) bool {
	return u.Active() == unit.Active
}
func isActivating(u Supervisable) bool {
	return u.Active() == unit.Activating
}
func (sys *System) isLoaded(name string) (loaded bool) {
	_, loaded = sys.Loaded[name]
	return
}
