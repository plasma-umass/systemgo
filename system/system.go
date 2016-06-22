package system

import (
	"io/ioutil"
	"time"

	"github.com/b1101/systemgo/unit"
)

var DEFAULT_PATHS = [...]string{"/etc/systemd/system/", "/run/systemd/system", "/lib/systemd/system"}

type System struct {
	// Map containing all loaded units
	loaded map[string]*Unit

	// Map containing all parsed units(includes units failed to load)
	parsed map[string]*Unit

	// Map containing all parsed paths(includes units specified as symlinks)
	// in *.wants and *.required directories
	parsedPaths map[string]*Unit

	// Slice of units in the queue
	queue *Queue

	// Status of global state
	state State

	// Deal with concurrency
	//sync.Mutex

	// Paths, where the unit file specifications get searched for
	paths []string

	// Starting time
	since time.Time

	// System log
	log *Log
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
func New() (sys *System) {
	defer func() {
		go sys.queueStarter()
	}()
	return &System{
		since: time.Now(),
		queue: NewQueue(),
		log:   NewLog(),
		paths: DEFAULT_PATHS,
	}
}

func (sys *System) Start(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		sys.Queue.Add(u)
	}
	return
}
func (sys *System) Stop(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		sys.Queue.Remove(u)
		u.Stop()
	}
	return
}
func (sys *System) Restart(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		sys.Queue.Remove(u)
		u.Stop()
		sys.Queue.Add(u)
	}
	return
}
func (sys *System) Reload(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		if reloader, ok := u.Supervisable.(Reloader); ok {
			reloader.Reload()
		} else {
			err = errors.NotFound
		}
	}
	return
}
func (sys *System) Enable(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		sys.Enabled[u] = true
	}
	return
}
func (sys *System) Disable(name string) (err error) {
	var u *Unit
	if u, err = sys.unit(name); err == nil {
		delete(sys.Enabled, u)
	}
	return
}

//func (sys System) WriteStatus(output io.Writer, names ...string) (err error) {
//if len(names) == 0 {
//w := tabwriter.Writer
//out += fmt.Sprintln("unit\t\t\t\tload\tactive\tsub\tdescription")
//out += fmt.Sprintln(s.Units)
//}

func (sys System) Status() (st Status, err error) {
	st = Status{
		State:  sys.State,
		Jobs:   sys.Queue.Len(),
		Failed: len(sys.Failed),
		Since:  sys.Since,
	}

	st.Log, err = ioutil.ReadAll(sys.Log)

	return

}
func (sys System) StatusOf(name string) (st unit.Status, err error) {
	var u *Unit
	if u, err = sys.unit(name); err != nil {
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
