package system

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/rvolosatovs/systemgo/unit"
)

var ErrIsStarting = errors.New("Unit is already starting")

type Unit struct {
	unit.Interface

	// System the Unit came from
	System *Daemon

	// Unit log
	Log *Log

	name string
	path string
	load unit.Load

	job *job

	mutex sync.Mutex
}

// TODO introduce a better workaround
const (
	starting  = "starting"
	stopping  = "stopping"
	reloading = "reloading"
)

// NewUnit returns an instance of new unit wrapping v
func NewUnit(v unit.Interface) (u *Unit) {
	return &Unit{
		Interface: v,
		Log:       NewLog(),
	}
}

//func (u *Unit) String() string {
//return u.Name()
//}

// Path returns path to the defintion unit was loaded from
func (u *Unit) Path() string {
	return u.path
}

// Name returns the name of the unit(filename of the defintion)
func (u *Unit) Name() string {
	return u.name
}

// Loaded returns load state of the unit
func (u *Unit) Loaded() unit.Load {
	return u.load
}

func (u *Unit) IsDead() bool {
	return u.Active() == unit.Inactive
}
func (u *Unit) IsActive() bool {
	return u.Active() == unit.Active
}
func (u *Unit) IsActivating() bool {
	return u.Active() == unit.Activating
}
func (u *Unit) IsDeactivating() bool {
	return u.Active() == unit.Deactivating
}
func (u *Unit) IsReloading() bool {
	return u.Active() == unit.Reloading
}

func (u *Unit) IsLoaded() bool {
	return u.Loaded() == unit.Loaded
}

// IsReloader returns whether u.Interface is capable of reloading
func (u *Unit) IsReloader() (ok bool) {
	_, ok = u.Interface.(unit.Reloader)
	return
}

func (u *Unit) Active() (st unit.Activation) {
	if u.jobRunning() {
		switch u.job.typ {
		case start:
			return unit.Activating
		case stop:
			return unit.Deactivating
		case reload:
			return unit.Reloading
		}
	}

	return u.Interface.Active()
}

func (u *Unit) Sub() string {
	if u.jobRunning() {
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

func (u *Unit) jobRunning() bool {
	return u.job != nil && u.job.IsRunning()
}

// Status returns status of the unit
func (u *Unit) Status() unit.Status {
	st := unit.Status{
		Load: unit.LoadStatus{
			Path:   u.Path(),
			Loaded: u.Loaded(),
			State:  -1, // TODO
		},
		Activation: unit.ActivationStatus{
			State: u.Active(),
			Sub:   u.Sub(),
		},
	}

	var err error
	if st.Log, err = ioutil.ReadAll(u.Log); err != nil {
		u.Log.Errorf("Error reading log: %s", err)
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

// Enable creates symlinks to u definition in dependency directories of each unit dependant on u
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
	if err = os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	return os.Symlink(dep.Path(), filepath.Join(dir, dep.Name()))
}

// Disable removes symlinks(if they exist) created by Enable
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
	if err = os.Remove(filepath.Join(dir, dep.Name())); err != nil && !os.IsNotExist(err) {
		return
	}
	return nil
}

// Reload creates a new reload transaction and runs it
func (u *Unit) Reload() (err error) {
	log.WithField("u", u).Debugf("u.Reload")

	tr := newTransaction()
	if err = tr.add(reload, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) reload() (err error) {
	log.WithField("u", u).Debugf("u.reload")

	reloader, ok := u.Interface.(unit.Reloader)
	if !ok {
		return ErrNoReload
	}
	return reloader.Reload()
}

// Start creates a new start transaction and runs it
func (u *Unit) Start() (err error) {
	log.WithField("unit", u.Name()).Debugf("u.Start")

	tr := newTransaction()
	if err = tr.add(start, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) start() (err error) {
	e := log.WithField("unit", u.Name())
	e.Debugf("u.start")

	if !u.IsLoaded() {
		e.Debug("not loaded")
		return ErrNotLoaded
	}

	u.Log.Println("Starting...")

	starter, ok := u.Interface.(unit.Starter)
	if !ok {
		e.Debugf("Interface is not unit.Starter")
		return nil
	}

	e.Debugf("Interface.Start")
	return starter.Start()
}

// Stop creates a new stop transaction and runs it
func (u *Unit) Stop() (err error) {
	log.WithField("u", u).Debugf("u.Stop")

	tr := newTransaction()
	if err = tr.add(stop, u, nil, true, true); err != nil {
		return
	}
	return tr.Run()
}

func (u *Unit) stop() (err error) {
	log.WithField("u", u).Debugf("u.stop")

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
