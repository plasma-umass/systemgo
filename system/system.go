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
		Log:    NewLog(),
		paths:  DEFAULT_PATHS,
		units:  make(map[string]*Unit),
		loaded: make(map[string]*Unit),
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
	if u, err = sys.Unit(name); err != nil {
		u, err = sys.Load(name)
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

// Status returns status of the system
func (sys *System) Status() (st Status, err error) {
	st = Status{
		State: sys.state,
		Since: sys.since,
	}

	st.Log, err = ioutil.ReadAll(sys.Log)

	return
}

// Load searches for a definition of unit name in configured paths parses it and returns a unit.Supervisable (or nil) and error if any
func (sys *System) Load(name string) (u *Unit, err error) {
	if !SupportedName(name) {
		return nil, ErrUnknownType
	}

	var paths []string
	if filepath.IsAbs(name) {
		paths = []string{name}
	} else {
		paths = make([]string, len(sys.paths))
		for i, path := range sys.paths {
			paths[i] = filepath.Join(path, name)
		}
	}

	for _, path := range paths {
		var sup Supervisable
		if sup, err = parseFile(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
		}

		var loaded bool
		if u, loaded = sys.loaded[name]; loaded {
			u.Supervisable = sup
		} else {
			u = sys.NewUnit(sup)
			sys.loaded[name] = u
		}
		u.path = path

		if err != nil {
			u.Log.Printf("Error parsing %s: %s", path, err)
			u.loaded = unit.Error
			return
		}

		u.loaded = unit.Loaded
		sys.units[name] = u
		return
	}

	return nil, ErrNotFound
}
