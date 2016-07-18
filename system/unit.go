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

var ErrIsLoading = errors.New("Unit is already loading")

type Unit struct {
	unit.Interface

	Log *Log

	path   string
	loaded unit.Load

	// Interfaces are too expensive to use?
	//requires map[string]activeWaiter
	requires map[string]*Unit

	loading chan struct{}
}

//type activeWaiter interface {
//IsActive() bool
//Wait()
//}

func NewUnit(v unit.Interface) (u *Unit) {
	if debug {
		defer func() {
			u.Log.Hooks.Add(&errorHook{fmt.Sprintf("%p", u)})
		}()
	}
	return &Unit{
		Interface: v,
		Log:       NewLog(),
		requires:  map[string]*Unit{},
	}
}

func (u *Unit) Path() string {
	return u.path
}
func (u *Unit) Loaded() unit.Load {
	return u.loaded
}

func (u *Unit) Status() fmt.Stringer {
	st := unit.Status{
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
	// TODO deal with different unit types requiring different status
	// something like u.Interface.HasX() ?
	switch u.Interface.(type) {
	//case *unit.Service:
	//return unit.ServiceStatus{st}
	default:
		return st
	}
}

func (u *Unit) Active() unit.Activation {
	if u.Interface == nil {
		return unit.Inactive
	} else {
		return u.Interface.Active()
	}
}

func (u *Unit) Sub() string {
	if u.Interface == nil {
		return "dead"
	} else {
		return u.Interface.Sub()
	}
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

func (u *Unit) Start() (err error) {
	defer func() {
		if err != nil {
			log.Debugf("%p failed to start, error: %s", u, err)
		} else {
			log.Debugf("%p started successfully", u)
		}
	}()

	log.Debugf("Start called on %p", u)
	if u.IsLoading() {
		return ErrIsLoading
	}

	u.Log.Println("Starting...")

	u.loading = make(chan struct{})
	defer func() {
		log.Debugf("%p finished loading", u)
		close(u.loading)
		u.loading = nil
	}()

	wg := &sync.WaitGroup{}
	for name, dep := range u.requires {
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
		return
	}

	return u.Interface.Start()
}
func (u *Unit) Stop() (err error) {
	if u.Interface == nil {
		return ErrNotLoaded
	}

	return u.Interface.Stop()
}
func (u *Unit) Wait() {
	if u.loading == nil {
		return
	}
	<-u.loading
	return
}

func (u *Unit) IsActive() bool {
	return u.Active() == unit.Active
}
func (u *Unit) IsLoading() bool {
	return u.loading != nil
}

func (u *Unit) parseDepDir(suffix string) (paths []string, err error) {
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
