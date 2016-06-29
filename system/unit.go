package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	Supervisable

	Log *Log

	path   string
	loaded unit.Load

	system *System

	loading chan struct{}
}

func (sys *System) NewUnit(sup Supervisable) (u *Unit) {
	return &Unit{
		Supervisable: sup,
		Log:          NewLog(),

		system:  sys,
		loading: make(chan struct{}),
	}
}

func (u Unit) Path() string {
	return u.path
}
func (u Unit) Loaded() unit.Load {
	return u.loaded
}
func (u Unit) Description() string {
	if u.Supervisable == nil {
		return ""
	}

	return u.Supervisable.Description()
}

func (u Unit) Active() unit.Activation {
	if u.Supervisable == nil {
		return unit.Inactive
	}

	if u.loading != nil {
		return unit.Activating
	}

	if subber, ok := u.Supervisable.(unit.Subber); ok {
		return subber.Active()
	}

	for _, name := range u.Requires() {
		if dep, err := u.system.Get(name); err != nil || !dep.isActive() {
			return unit.Inactive
		}
	}

	return unit.Active
}

func (u Unit) Sub() string {
	if u.Supervisable == nil {
		return "dead"
	}

	if subber, ok := u.Supervisable.(unit.Subber); ok {
		return subber.Sub()
	}

	return u.Active().String()
}

func (u *Unit) Requires() (names []string) {
	if u.Supervisable != nil {
		names = u.Supervisable.Requires()
	}

	if paths, err := u.parseDepDir(".requires"); err == nil {
		names = append(names, paths...)
	}

	return
}

func (u *Unit) Wants() (names []string) {
	if u.Supervisable != nil {
		names = u.Supervisable.Wants()
	}

	if paths, err := u.parseDepDir(".wants"); err == nil {
		names = append(names, paths...)
	}

	return
}

func (u *Unit) parseDepDir(suffix string) (paths []string, err error) {
	dirpath := u.path + suffix

	links, err := pathset(dirpath)
	if err != nil {
		if err != os.ErrNotExist {
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

func (u *Unit) Start() (err error) {
	u.loading = make(chan struct{})
	defer close(u.loading)

	u.Log.Println("Starting unit...")

	// TODO: stop conflicted units before starting(divide jobs and use transactions like systemd?)
	u.Log.Println("Checking Conflicts...")
	for _, name := range u.Conflicts() {
		if dep, _ := u.system.Get(name); dep != nil && dep.isActive() {
			return fmt.Errorf("Unit conflicts with %s", name)
		}
	}

	u.Log.Println("Checking Requires...")
	for _, name := range u.Requires() {
		var dep *Unit
		if dep, err = u.system.Get(name); err != nil {
			return fmt.Errorf("Error loading dependency %s: %s", name, err)
		}

		if !dep.isActive() {
			dep.waitFor()
			if !dep.isActive() {
				return fmt.Errorf("Dependency %s failed to start", name)
			}
		}
	}

	if u.Supervisable == nil {
		return ErrNotLoaded
	}

	if starter, ok := u.Supervisable.(unit.Starter); ok {
		err = starter.Start()
	}
	return
}

func (u *Unit) Stop() (err error) {
	return ErrNotImplemented
}
func (u Unit) isActive() bool {
	return u.Active() == unit.Active
}
func (u Unit) isLoading() bool {
	return u.Active() == unit.Activating
}
func (u Unit) waitFor() {
	<-u.loading
	return
}
