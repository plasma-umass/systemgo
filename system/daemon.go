package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rvolosatovs/systemgo/unit"
	"github.com/rvolosatovs/systemgo/unit/service"
	"github.com/rvolosatovs/systemgo/unit/target"

	log "github.com/Sirupsen/logrus"
)

var ErrDepConflict = fmt.Errorf("Error stopping conflicting unit")
var ErrNotActive = fmt.Errorf("Unit is not active")
var ErrExists = fmt.Errorf("Unit already exists")

var DEFAULT_PATHS = []string{"/etc/systemd/system/", "/run/systemd/system", "/lib/systemd/system"}

type Daemon struct {
	// Map of created units (name -> *Unit)
	units map[string]*Unit

	// Paths, where the unit file specifications get searched for
	paths []string

	// System state
	state State

	// Starting time
	since time.Time

	// System log
	Log *Log

	mutex sync.Mutex
}

func New() (sys *Daemon) {
	defer func() {
		if debug {
			sys.Log.Logger.Hooks.Add(&errorHook{
				Source: "system",
			})
		}
	}()
	return &Daemon{
		units: make(map[string]*Unit),

		since: time.Now(),
		Log:   NewLog(),
		paths: DEFAULT_PATHS,
	}
}

func (sys *Daemon) Paths() (paths []string) {
	return sys.paths
}

func (sys *Daemon) SetPaths(paths ...string) {
	sys.paths = paths
}

func (sys *Daemon) Since() (t time.Time) {
	return sys.since
}

// Status returns status of the system
// If error is returned it is going to be an error,
// returned by the call to ioutil.ReadAll(sys.Log)
func (sys *Daemon) Status() (st Status, err error) {
	st = Status{
		State: sys.state,
		Since: sys.since,
	}

	st.Log, err = ioutil.ReadAll(sys.Log)

	return
}

var supported = map[string]bool{
	".service": true,
	".target":  true,
	".mount":   false,
	".socket":  false,
}

// SupportedSuffix returns a bool indicating if suffix represents a unit type,
// which is supported by Systemgo
func SupportedSuffix(suffix string) bool {
	return supported[suffix]
}

// Supported returns a bool indicating if filename represents a unit type,
// which is supported by Systemgo
func Supported(filename string) bool {
	return SupportedSuffix(filepath.Ext(filename))
}

// IsEnabled returns enable state of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *Daemon) IsEnabled(name string) (st unit.Enable, err error) {
	//var u *Unit
	//if u, err = sys.Unit(name); err == nil && sys.Enabled[u] {
	//st = unit.Enabled
	//}
	return unit.Enabled, ErrNotImplemented
}

// IsActive returns activation state of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *Daemon) IsActive(name string) (st unit.Activation, err error) {
	var u *Unit
	if u, err = sys.Get(name); err == nil {
		st = u.Active()
	}
	return
}

// StatusOf returns status of the unit held in-memory under specified name
// If error is returned, it is going to be ErrNotFound
func (sys *Daemon) StatusOf(name string) (st unit.Status, err error) {
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
}

func (sys *Daemon) failedCount() (n int) {
	for _, u := range sys.units {
		if j := u.job; j != nil && j.Failed() {
			n++
		}
	}
	return
}

func (sys *Daemon) jobCount() (n int) {
	for _, u := range sys.units {
		if u.job != nil {
			n++
		}
	}
	return
}

//func (sys Daemon) WriteStatus(output io.Writer, names ...string) (err error) {
//if len(names) == 0 {
//w := tabwriter.Writer
//out += fmt.Sprintln("unit\t\t\t\tload\tactive\tsub\tdescription")
//out += fmt.Sprintln(s.Units)
//}

//func (us units) String() (out string) {
//for _, u := range us {
//out += fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t\n",
//u.Name(), u.Loaded(), u.Active(), u.Sub(), u.Description())
//}
//return
//}

//func (sys *Daemon) newUnit(v unit.Interface) (u *Unit) {
//	u = NewUnit(v)
//	sys.Supervise(u)
//}

func (sys *Daemon) Supervise(name string, v unit.Interface) (u *Unit, err error) {
	if _, exists := sys.units[name]; exists {
		return nil, ErrExists
	}

	u = NewUnit(v)
	u.System = sys

	u.name = name
	sys.units[name] = u

	log.WithFields(log.Fields{
		"unit": name,
	}).Debugf("Created new *Unit")

	return
}

// Unit looks up the unit name in the internal hasmap of loaded units and returns it
// If error is returned, it will be error from sys.Load(name)
func (sys *Daemon) Unit(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.units[name]; !ok {
		return nil, ErrNotFound
	}
	return
}

// Get looks up the unit name in the internal hasmap of loaded units and calls
// sys.Load(name) if it can not be found
// If error is returned, it will be error from sys.Load(name)
func (sys *Daemon) Get(name string) (u *Unit, err error) {
	if u, err = sys.Unit(name); err != nil || !u.IsLoaded() {
		return sys.Load(name)
	}
	return
}

func (sys *Daemon) getAndExecute(names []string, fn func(*Unit) error) (err error) {
	for _, name := range names {
		var u *Unit
		if u, err = sys.Get(name); err != nil {
			return
		}

		if err = fn(u); err != nil {
			return
		}
	}
	return
}

// TODO
func (sys *Daemon) Enable(names ...string) (err error) {
	return ErrNotImplemented
}

// TODO
func (sys *Daemon) Disable(names ...string) (err error) {
	return ErrNotImplemented
}

// TODO: isolate, check redundant jobs, fail fast

func (sys *Daemon) Stop(names ...string) (err error) {
	var tr *transaction
	if tr, err = sys.newTransaction(stop, names); err != nil {
		return
	}
	return tr.Run()
}

func (sys *Daemon) Start(names ...string) (err error) {
	var tr *transaction
	if tr, err = sys.newTransaction(start, names); err != nil {
		return
	}
	return tr.Run()
}

func (sys *Daemon) Restart(names ...string) (err error) {
	var tr *transaction
	if tr, err = sys.newTransaction(restart, names); err != nil {
		return
	}
	return tr.Run()
}

func (sys *Daemon) newTransaction(t jobType, names []string) (tr *transaction, err error) {
	sys.mutex.Lock()
	defer sys.mutex.Unlock()

	tr = newTransaction()

	for _, name := range names {
		var dep *Unit
		if dep, err = sys.Get(name); err != nil {
			return nil, err
		}

		if err = tr.add(t, dep, nil, true, false); err != nil {
			return nil, err
		}
	}
	return
}

// Load searches for name in configured paths, parses it, and either overwrites the definition of already
// created Unit or creates a new one
func (sys *Daemon) Load(name string) (u *Unit, err error) {
	log.WithField("name", name).Debugln("Load")

	var parsed bool
	u, parsed = sys.units[name]

	if !Supported(name) {
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
		var file *os.File
		if file, err = os.Open(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		defer file.Close()

		if !parsed {
			if u, err = sys.parse(path); err != nil {
				return
			}

			if name != path {
				sys.units[path] = u
			}
		}

		u.path = path

		var info os.FileInfo
		if info, err = file.Stat(); err == nil && info.IsDir() {
			err = ErrIsDir
		}
		if err != nil {
			u.Log.Printf("%s", err)
			return u, err
		}

		if err = u.Interface.Define(file); err != nil {
			if me, ok := err.(unit.MultiError); ok {
				u.Log.Printf("Definition is invalid:")
				for _, errmsg := range me.Errors() {
					u.Log.Printf(errmsg)
				}
			} else {
				u.Log.Printf("Error parsing definition: %s", err)
			}
			u.loaded = unit.Error
			return u, err
		}

		u.loaded = unit.Loaded

		return u, err
	}

	return nil, ErrNotFound
}

func (sys *Daemon) parse(name string) (u *Unit, err error) {
	var v unit.Interface
	switch filepath.Ext(name) {
	case ".target":
		v = &target.Unit{}
	case ".service":
		v = &service.Unit{}
	default:
		panic("Trying to load an unsupported unit type")
	}

	return sys.Supervise(name, v)
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
			definitions = append(definitions, filepath.Join(path, name))
		}
	}

	return
}
