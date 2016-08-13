package system

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/rvolosatovs/systemgo/unit"
)

var ErrIsStarting = errors.New("Unit is already starting")

type Unit struct {
	unit.Interface

	name string
	path string

	loaded unit.Load

	Log *Log

	System *Daemon

	mutex sync.Mutex

	job *job
}

// TODO introduce a better workaround
const (
	starting  = "starting"
	stopping  = "stopping"
	reloading = "reloading"
)

func NewUnit(v unit.Interface) (u *Unit) {
	return &Unit{
		Interface: v,
		Log:       NewLog(),
	}
}

func (u *Unit) Path() string {
	return u.path
}
func (u *Unit) Name() string {
	return u.name
}
func (u *Unit) Loaded() unit.Load {
	return u.loaded
}

func (u *Unit) IsActive() bool {
	return u.Active() == unit.Active
}
func (u *Unit) IsStarting() bool {
	return u.Active() == unit.Activating
}
func (u *Unit) IsLoaded() bool {
	return u.Loaded() == unit.Loaded
}

func (u *Unit) Active() (st unit.Activation) {
	switch u.Sub() {
	case starting:
		return unit.Activating
	case stopping:
		return unit.Deactivating
	case reloading:
		return unit.Reloading
	default:
		return u.Interface.Active()
	}
}

func (u *Unit) Sub() string {
	if u.job != nil && u.job.IsRunning() {
		switch u.job.typ {
		case start:
			return starting
		case stop:
			return stopping
		case reload:
			return reloading
		}
	}

	return u.Interface.Sub()
}

func (u *Unit) Status() fmt.Stringer {
	st := unit.Status{
		Load: unit.LoadStatus{
			Path:   u.Path(),
			Loaded: u.Loaded(),
			State:  unit.Enabled, // TODO
		},
		Activation: unit.ActivationStatus{
			State: u.Active(),
			Sub:   u.Sub(),
		},
	}
	// TODO deal with different unit types requiring different status
	// something like u.Interface.HasX() ?
	switch u.Interface.(type) {
	//case *unit.Service:
	//return unit.ServiceStatus{st}
	default:
		return st
	}
}

func (u *Unit) IsReloader() (ok bool) {
	_, ok = u.Interface.(unit.Reloader)
	return
}

func (u *Unit) Reload() (err error) {
	// TODO reload transaction
	return
}

// Requires returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Requires() (names []string) {
	names = u.Interface.Requires()

	if paths, err := readDepDir(u.requiresDir()); err == nil {
		names = append(names, paths...)
	}

	return
}

// Wants returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Wants() (names []string) {
	names = u.Interface.Wants()

	if paths, err := readDepDir(u.wantsDir()); err == nil {
		names = append(names, paths...)
	}

	return
}

func (u *Unit) wantsDir() (path string) {
	return u.depDir("wants")
}

func (u *Unit) requiresDir() (path string) {
	return u.depDir("requires")
}

func (u *Unit) depDir(suffix string) (path string) {
	return u.Path() + "." + suffix
}

func (u *Unit) Enable() (err error) {
	err = u.System.getAndExecute(u.RequiredBy(), func(dep *Unit, gerr error) error {
		if gerr != nil {
			return gerr
		}

		return dep.addRequiresDep(u)
	})
	if err != nil {
		return
	}

	return u.System.getAndExecute(u.WantedBy(), func(dep *Unit, gerr error) error {
		if gerr != nil {
			return gerr
		}

		return dep.addWantsDep(u)
	})
}

func (u *Unit) addWantsDep(dep *Unit) (err error) {
	return linkDep(u.wantsDir(), dep)
}

func (u *Unit) addRequiresDep(dep *Unit) (err error) {
	return linkDep(u.requiresDir(), dep)
}

func linkDep(dir string, dep *Unit) (err error) {
	if err = os.Mkdir(dir, 0755); err != nil && err != os.ErrExist {
		return err
	}

	return os.Symlink(dep.Path(), filepath.Join(dir, dep.Name()))
}

func (u *Unit) Disable() (err error) {
	err = u.System.getAndExecute(u.RequiredBy(), func(dep *Unit, gerr error) error {
		if gerr != nil {
			return gerr
		}

		return dep.removeRequiresDep(u)
	})
	if err != nil {
		return
	}

	return u.System.getAndExecute(u.WantedBy(), func(dep *Unit, gerr error) error {
		if gerr != nil {
			return gerr
		}

		return dep.removeWantsDep(u)
	})
}

func (u *Unit) removeWantsDep(dep *Unit) (err error) {
	return unlinkDep(u.wantsDir(), dep)
}

func (u *Unit) removeRequiresDep(dep *Unit) (err error) {
	return unlinkDep(u.requiresDir(), dep)
}

func unlinkDep(dir string, dep *Unit) (err error) {
	if err = os.Remove(filepath.Join(dir, dep.Name())); err != nil && err != os.ErrNotExist {
		return
	}
	return nil
}

func (u *Unit) reload() (err error) {
	reloader, ok := u.Interface.(unit.Reloader)
	if !ok {
		return ErrNoReload
	}
	return reloader.Reload()
}

func (u *Unit) Start() (err error) {
	tr := newTransaction()
	if err = tr.add(start, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) start() (err error) {
	log.Debugf("start called on %v", u)

	if !u.IsLoaded() {
		return ErrNotLoaded
	}

	u.Log.Println("Starting...")

	starter, ok := u.Interface.(unit.Starter)
	if !ok {
		return nil
	}

	return starter.Start()
}

func (u *Unit) Stop() (err error) {
	tr := newTransaction()
	if err = tr.add(stop, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) stop() (err error) {
	log.Debugf("stop called on %v", u)

	if !u.IsLoaded() {
		return ErrNotLoaded
	}

	u.Log.Println("Stopping...")

	stopper, ok := u.Interface.(unit.Stopper)
	if !ok {
		return nil
	}

	return stopper.Stop()
}

func readDepDir(dir string) (paths []string, err error) {
	var links []string
	if links, err = pathset(dir); err != nil {
		return
	}

	paths = make([]string, 0, len(links))
	for _, path := range links {
		if path, err = filepath.EvalSymlinks(path); err != nil {
			return
		}
		paths = append(paths, path)
	}
	return
}
