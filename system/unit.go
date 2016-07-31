package system

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/b1101/systemgo/unit"
)

var ErrIsStarting = errors.New("Unit is already starting")

type Unit struct {
	unit.Interface

	loading

	name string
	path string

	loaded unit.Load

	Log *Log

	System *Daemon

	mutex sync.Mutex

	job *job
}

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

func (u *Unit) Active() unit.Activation {
	switch {
	case u.job != nil:
		switch u.job.typ {
		case start:
			return "starting"
		case stop:
			return "stopping"
		}
	case u.Interface == nil:
		return unit.Inactive
	default:
		return u.Interface.Active()
	}
}

func (u *Unit) Sub() string {
	if u.Interface == nil {
		return "dead"
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
	reloader, ok := u.Interface.(unit.Reloader)
	if !ok {
		return ErrNoReload
	}
	return reloader.Reload()
}

type loading chan struct{}

func (l *loading) Start() {
	*l = make(loading)
}

func (l *loading) Stop() {
	close(*l)
}

func (l *loading) Wait() {
	<-*l
}

func (u *Unit) Wait() {
	u.loading.Wait()
}

func (u *Unit) start() (err error) {
	t := newTransaction()
	j := &startJob{
		u,
	}
	t.add(j)
	t.anchor(j)
}

func (u *Unit) Start() (err error) {
	log.Debugf("Start called on %p", u)

	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.loading.Start()
	defer u.loading.Stop()

	if !u.IsLoaded() {
		return ErrNotLoaded
	} else if u.IsStarting() {
		return ErrIsStarting
	}

	u.Log.Println("Starting...")

	wg := &sync.WaitGroup{}
	//for name, dep := range u.requires {
	for _, name := range u.Requires() {
		var dep *Unit
		if dep, err = u.system.Get(name); err != nil {
			u.Log.Printf("Failed to load dependency %s: %s", name, err)
			return ErrDepFail
		}
		wg.Add(1)
		go func(name string, dep *Unit) {
			defer wg.Done()
			log.Debugf("%p waiting for %p(%s)", u, dep, name)
			dep.Wait()
			log.Debugf("%p finished waiting for %p(%s)", u, dep, name)
			if !dep.IsActive() {
				u.Log.Printf("Dependency %s failed to start", name)
				err = ErrDepFail
			}
		}(name, dep)
	}

	wg.Wait()
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

func (u *Unit) Stop() (err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.loading.Start()
	defer u.loading.Stop()

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
