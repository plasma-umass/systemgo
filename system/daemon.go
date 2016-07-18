package system

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/b1101/systemgo/unit"
	"github.com/b1101/systemgo/unit/service"
	"github.com/b1101/systemgo/unit/target"

	log "github.com/Sirupsen/logrus"
)

var ErrDepConflict = fmt.Errorf("Error stopping conflicting unit")

var DEFAULT_PATHS = []string{"/etc/systemd/system/", "/run/systemd/system", "/lib/systemd/system"}

type Daemon struct {
	// Map containing pointers to all currently active units(name -> *Unit)
	active map[*Unit]bool

	// Map containing pointers to all successfully loaded units(name -> *Unit)
	loaded map[*Unit]bool

	// Map of created units (name -> *Unit)
	units map[string]*Unit

	// Paths, where the unit file specifications get searched for
	Paths []string

	// System state
	State State

	// Starting time
	Since time.Time

	// System log
	Log *Log

	Jobs Jobs
}

type Jobs struct {
	sync.Mutex

	assigned map[*Unit]job

	failed map[*Unit]bool

	units chan *Unit
}

func (jobs *Jobs) Len() (n int) {
	return len(jobs.assigned)
}

func (jobs *Jobs) Failed() (n int) {
	return len(jobs.failed)
}

func (jobs *Jobs) Assign(u *Unit, j job) {
	jobs.Lock()
	defer jobs.Unlock()

	assigned, has := jobs.assigned[u]
	if !has {
		log.Debugf("Assigned a new job(%v) for %p", j, u)
		jobs.assigned[u] = j
		jobs.units <- u
		return
	}

	log.Debugf("A job for %p has already been assigned", u)

	switch {
	case assigned == stop && j == start:
		jobs.assigned[u] = restart

	case assigned == start && j == stop:
		delete(jobs.assigned, u)

	default:
		jobs.assigned[u] = j
	}
}

func (sys *Daemon) doJobs() {
	for u := range sys.Jobs.units {
		log.Debugf("sys.work() unit: %p", u)

		go func(u *Unit) {
			sys.Jobs.Lock()
			defer sys.Jobs.Unlock()

			if j, has := sys.Jobs.assigned[u]; has {
				if err := sys.do(u, j); err != nil {
					log.Debugf("Job for %p failed to execute: %s", u, err)
					sys.Jobs.failed[u] = true
				}
				delete(sys.Jobs.assigned, u)
			}
		}(u)
	}
}

func (sys *Daemon) do(u *Unit, j job) (err error) {
	switch j {
	case start:
		if err = u.Start(); err == nil {
			sys.active[sys.nameOf(u)] = u
		}
		return
	case stop:
		if err = u.Stop(); err == nil {
			delete(sys.active, sys.nameOf(u))
		}
		return
	case restart:
		if err = sys.do(u, stop); err == nil {
			err = sys.do(u, start)
		}
		return
	default:
		panic("Unknown job type")
	}
}

type job int

const (
	start job = iota
	stop
	restart
)

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

func New() (sys *Daemon) {
	defer func() {
		go sys.doJobs()

		if debug {
			sys.Log.Logger.Hooks.Add(&errorHook{
				Source: "system",
			})
		}
	}()
	return &Daemon{
		active: make(map[string]*Unit),
		loaded: make(map[string]*Unit),
		parsed: make(map[string]*Unit),
		names:  make(map[*Unit]string),

		Since: time.Now(),
		Log:   NewLog(),
		Paths: DEFAULT_PATHS,

		Jobs: Jobs{
			units:    make(chan *Unit),
			assigned: map[*Unit]job{},
		},
	}
}

func (sys *Daemon) SetPaths(paths ...string) {
	sys.Paths = paths
}

// Status returns status of the system
// If error is returned it is going to be an error,
// returned by the call to ioutil.ReadAll(sys.Log)
func (sys *Daemon) Status() (st Status, err error) {
	st = Status{
		State: sys.State,
		Since: sys.Since,
	}

	st.Log, err = ioutil.ReadAll(sys.Log)

	return
}

func (sys *Daemon) Start(names ...string) (err error) {
	log.Debugf("sys.Start names:\n%+v", names)

	var units map[string]*Unit
	if units, err = sys.loadDeps(names); err != nil {
		return
	}
	log.Debugf("sys.loadDeps returned:\n%+v, nil", units)

	var ordering []*Unit
	if ordering, err = sys.order(units); err != nil {
		return
	}
	log.Debugf("sys.order returned:\n%+v, nil", ordering)

	for _, u := range ordering {
		wg := &sync.WaitGroup{}
		for _, name := range u.Conflicts() {
			log.Debugf("%p conflicts with %s", u, name)
			wg.Add(1)
			go func() {
				if err = sys.Stop(name); err != nil && err != ErrNotFound {
					u.Log.Printf("Error stopping %s: %s", name, err)
				}
				wg.Done()
			}()

		}
		wg.Wait()
		if err != nil {
			return ErrDepConflict
		}

		sys.Jobs.Assign(u, start)
	}

	return
}

func (sys *Daemon) Stop(name string) (err error) {
	log.Debugf("sys.Stop name: %s", name)

	if !sys.active[u] {
		return ErrNotActive
	}
	return
}

func (sys *Daemon) Restart(name string) (err error) {
	if u, ok := sys.active[name]; ok {
		sys.Jobs.Assign(u, restart)
	}
	return
}

func (sys *Daemon) Reload(name string) (err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
		return
	}

	if reloader, ok := u.Interface.(unit.Reloader); ok {
		return reloader.Reload()
	}

	return ErrNoReload
}

// TODO
func (sys *Daemon) Enable(name string) (err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
		return
	}
	u.Log.Println("enable")
	return ErrNotImplemented
}

// TODO
func (sys *Daemon) Disable(name string) (err error) {
	var u *Unit
	if u, err = sys.Get(name); err != nil {
		return
	}
	u.Log.Println("disable")
	return ErrNotImplemented
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

var std = New()

// Get looks up the unit name in the internal hasmap of loaded units and calls
// sys.Load(name) if it can not be found
// If error is returned, it will be error from sys.Load(name)
func (sys *Daemon) Get(name string) (u *Unit, err error) {
	var ok bool
	if u, ok = sys.units[name]; !ok {
		u, err = sys.Load(name)
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

// Load searches for a definition of unit name in configured paths parses it and returns a pointer to Unit
// If a unit name has already been parsed(tried to load) by sys, it will not create a new unit, but return a pointer to that unit instead
func (sys *Daemon) Load(name string) (u *Unit, err error) {
	if !Supported(name) {
		return nil, ErrUnknownType
	}

	var paths []string
	if filepath.IsAbs(name) {
		paths = []string{name}
	} else {
		paths = make([]string, len(sys.Paths))
		for i, path := range sys.Paths {
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

		var parsed bool
		if u, parsed = sys.units[name]; !parsed {
			var v unit.Interface
			switch filepath.Ext(path) {
			case ".target":
				v = &target.Unit{Get: func(name string) (unit.Subber, error) {
					return sys.Get(name)
				}}
			case ".service":
				v = &service.Unit{}
			default:
				panic("Trying to load an unsupported unit type")
			}

			u = NewUnit(v)
			u.name = name
			log.WithFields(log.Fields{
				"unit": name,
			}).Debugf("Created new *Unit")

			sys.units[name] = u

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
		sys.loaded[u] = true

		return u, err
	}

	return nil, ErrNotFound
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

func (sys *Daemon) loadDeps(names []string) (units map[string]*Unit, err error) {
	log.Debugf("loadDeps names:\n%+v", names)

	units = map[string]*Unit{}
	added := func(name string) (is bool) {
		_, is = units[name]
		return
	}

	var failed bool
	for len(names) > 0 {
		name := names[0]

		if !added(name) {
			var u *Unit
			if u, err = sys.Get(name); err != nil {
				return nil, fmt.Errorf("Error loading dependency: %s", name)
			}
			units[name] = u

			names = append(names, u.Requires()...)

			for _, name := range u.Wants() {
				if !added(name) {
					units[name], _ = sys.Get(name)
				}
			}
		}

		names = names[1:]
	}
	if failed {
		return nil, ErrDepFail
	}

	return
}

type graph struct {
	ordered  map[*Unit]struct{}
	visited  map[*Unit]struct{}
	before   map[*Unit]map[string]*Unit
	ordering []*Unit
}

func (sys *Daemon) order(units map[string]*Unit) (ordering []*Unit, err error) {
	log.Debugf("sys.order recieved units:\n%+v", units)

	g := &graph{
		map[*Unit]struct{}{},
		map[*Unit]struct{}{},
		map[*Unit]map[string]*Unit{},
		make([]*Unit, 0, len(units)),
	}

	for _, unit := range units {
		g.before[unit] = map[string]*Unit{}
	}

	log.Debugf("g.before:\n%+v", g.before)

	for name, unit := range units {
		log.Debugf("Checking after of %s...", name)
		for _, depname := range unit.After() {
			log.Debugf("%s after %s", name, depname)
			if dep, ok := units[depname]; ok {
				g.before[unit][depname] = dep
			}
		}

		log.Debugf("Checking before of %s...", name)
		for _, depname := range unit.Before() {
			log.Debugf("%s before %s", name, depname)
			if dep, ok := units[depname]; ok {
				g.before[dep][name] = unit
			}
		}

		log.Debugf("Checking requires of %s...", name)
		for _, depname := range unit.Requires() {
			log.Debugf("added %s to %s.requires hashmap", depname, name)
			unit.requires[depname] = units[depname]
		}
	}

	log.Debugf("starting DFS on graph:\n%+v", g)
	for name, unit := range units {
		log.Debugf("picking %s", name)
		if err = g.traverse(unit); err != nil {
			return nil, fmt.Errorf("Dependency cycle determined:\n%s depends on %s", name, err)
		}
	}

	return g.ordering, nil
}

var errBlank = errors.New("")

func (g *graph) traverse(u *Unit) (err error) {
	log.Debugf("traverse %p, graph %p, ordering:\n%v", u, g, g.ordering)
	if _, has := g.ordered[u]; has {
		log.Debugf("%p already ordered", u)
		return nil
	}

	if _, has := g.visited[u]; has {
		log.Debugf("%p already visited", u)
		return errBlank
	}

	g.visited[u] = struct{}{}

	for name, dep := range g.before[u] {
		if err = g.traverse(dep); err != nil {
			if err == errBlank {
				return fmt.Errorf("%s\n", name)
			}
			return fmt.Errorf("%s\n%s depends on %s", name, name, err)
		}
	}

	delete(g.visited, u)

	if _, has := g.ordered[u]; !has {
		g.ordering = append(g.ordering, u)
		g.ordered[u] = struct{}{}
	}

	return nil
}
