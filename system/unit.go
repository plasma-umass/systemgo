package system

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	Supervisable

	Log *Log

	path   string
	loaded unit.Load

	Getter

	loading chan struct{}
}






func NewUnit(v Supervisable) (u *Unit) {
	return &Unit{
		Supervisable: v,
		Log:          NewLog(),

		system:  sys,
		loading: make(chan struct{}),
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
	switch u.Supervisable.(type) {
	//case *unit.Service:
	//return unit.ServiceStatus{st}
	default:
		return st
	}
}

func (u *Unit) Active() unit.Activation {
	if u.Supervisable == nil {
		return unit.Inactive
	} else {
		return u.Supervisable.Active()
	}
}

func (u *Unit) Sub() string {
	if u.Supervisable == nil {
		return "dead"
	} else {
		return u.Supervisable.Sub()
	}
}

// Description returns a string as found in Definition
func (u *Unit) Description() (description string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().Description()
}

// Documentation returns a string as found in definition
func (u *Unit) Documentation() (documentation string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().Documentation()
}

// Conflicts returns a slice of unit names as found in definition
func (u *Unit) Conflicts() (names []string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().Conflicts()
}

// After returns a slice of unit names as found in definition
func (u *Unit) After() (names []string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().After()
}

// Before returns a slice of unit names as found in definition
func (u *Unit) Before() (names []string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().Before()
}

// RequiredBy returns a slice of unit names as found in definition
func (u *Unit) RequiredBy() (names []string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().RequiredBy()
}

// WantedBy returns a slice of unit names as found in definition
func (u *Unit) WantedBy() (names []string) {
	if u.Supervisable == nil {
		return
	}
	return u.definition().WantedBy()
}

// Requires returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Requires() (names []string) {
	if u.Supervisable == nil {
		return
	}
	names = u.definition().Requires()

	if paths, err := u.parseDepDir(".requires"); err == nil {
		names = append(names, paths...)
	}

	return
}

// Wants returns a slice of unit names as found in definition and absolute paths
// of units symlinked in units '.wants' directory
func (u *Unit) Wants() (names []string) {
	if u.Supervisable != nil {
		names = u.definition().Wants()
	}

	if paths, err := u.parseDepDir(".wants"); err == nil {
		names = append(names, paths...)
	}

	return
}

// definition returns pointer to the definition of u
// Assumes that u has a "Definition" field, which implements definition
func (u *Unit) definition() (d definition) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln(r)
		}
	}()

	return reflect.ValueOf(u).Elem().FieldByName("Definition").Addr().Interface().(definition)
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

func (u *Unit) Start() (err error) {
	if Debug {
		bug.Println("*Unit is ", u)
	}

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
