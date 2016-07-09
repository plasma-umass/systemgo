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

func (sys *System) Get(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.units[name]; !ok {
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
	if u, err = sys.Get(name); err != nil {
		return
	}

	if reloader, ok := u.Supervisable.(unit.Reloader); ok {
		return reloader.Reload()
	}

	return ErrNoReload
}

func (sys *System) Enable(name string) (err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
		return
	}
	u.Log.Println("enable")
	return fmt.Errorf("TODO")
}

func (sys *System) Disable(name string) (err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
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
	if u, err = sys.Get(name); err == nil {
		st = u.Active()
	}
	return
}

// StatusOf returns status of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *System) StatusOf(name string) (st unit.Status, err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
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
// Supported returns a bool indicating if filename represents a unit type,
// which is supported by Systemgo
func Supported(filename string) bool {
	return SupportedSuffix(filepath.Ext(filename))
}

// Status returns status of the system
func (sys *System) Status() (st Status, err error) {
	st = Status{
		State: sys.state,
		Since: sys.since,
	}
var supported = map[string]bool{
	".service": true,
	".target":  true,
	".mount":   false,
	".socket":  false,
}

	st.Log, err = ioutil.ReadAll(sys.Log)
// SupportedSuffix returns a bool indicating if suffix represents a unit type,
// which is supported by Systemgo
func SupportedSuffix(suffix string) bool {
	return supported[suffix]
}

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
// pathset returns a slice of paths to definitions of supported unit types found in path specified
func pathset(path string) (definitions []string, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return nil, err
	}
	defer file.Close()

	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, ErrNotDir
	}

	var names []string
	if names, err = file.Readdirnames(0); err != nil {
		return nil, err
	}

	definitions = make([]string, 0, len(names))
	for _, name := range names {
		if Supported(name) {
			definitions = append(definitions, filepath.Clean(path+"/"+name))
		}
	}

	return
}
