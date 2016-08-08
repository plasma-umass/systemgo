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

// Requires returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Requires() (names []string) {
	names = u.Interface.Requires()

	if paths, err := u.parseDepDir(".requires"); err == nil {
		names = append(names, paths...)
	}

	return
}

// Wants returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Wants() (names []string) {
	names = u.Interface.Wants()

	if paths, err := u.parseDepDir(".wants"); err == nil {
		names = append(names, paths...)
	}

	return
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

func (u *Unit) reload() (err error) {
	reloader, ok := u.Interface.(unit.Reloader)
	if !ok {
		return ErrNoReload
	}
	return reloader.Reload()
}

func (u *Unit) start() (err error) {
	log.Debugf("start called on %v", u)

	if !u.IsLoaded() {
		return ErrNotLoaded
	}

	u.Log.Println("Starting...")

	log.Debugf("All dependencies of %p finished loading", u)
	if err != nil {
		return err
	}

	starter, ok := u.Interface.(unit.Starter)
	if !ok {
		return nil
	}

	return starter.Start()
}

func (u *Unit) Start() (err error) {
	tr := newTransaction()
	if err = tr.add(start, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) stop() (err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if !u.IsLoaded() {
		return ErrNotLoaded
	}

	stopper, ok := u.Interface.(unit.Stopper)
	if !ok {
		return nil
	}

	return stopper.Stop()
}

func (u *Unit) parseDepDir(suffix string) (paths []string, err error) {
	if u.path == "" {
		return paths, errors.New("Path is empty")
	}

	dirpath := u.path + suffix

	links, err := pathset(dirpath)
	if err != nil {
		if !os.IsNotExist(err) {
			u.Log.Printf("Error parsing %s: %s", dirpath, err)
		}
		return
	}

	paths = make([]string, 0, len(links))
	for _, path := range links {
		if path, err = filepath.EvalSymlinks(path); err != nil {
			u.Log.Printf("Error reading link at %s: %s", path, err)
			continue
		}
		paths = append(paths, path)
	}
	return
}
