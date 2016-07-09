package system

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/b1101/systemgo/unit"
)

var ErrIsLoading = errors.New("Unit is already loading")

type Unit struct {
	Supervisable

	Log *Log

	path   string
	loaded unit.Load

	requires map[string]dependency

	loading chan struct{}
}

type dependency interface {
	unit.Subber
	Wait()
}

func NewUnit(v Supervisable) (u *Unit) {
	return &Unit{
		Supervisable: v,
		Log:          NewLog(),
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

func (u *Unit) Start() (err error) {
	if u.Supervisable == nil {
		return ErrNotLoaded
	} else if u.isLoading() {
		return ErrIsLoading
	}

	u.Log.Println("Starting...")

	u.loading = make(chan struct{})
	defer func() {
		close(u.loading)
		u.loading = nil
	}()

	wg := &sync.WaitGroup{}
	for name, dep := range u.requires {
		name, dep := name, dep
		wg.Add(1)
		go func( /*name string, unit unit.Subber*/ ) {
			defer wg.Done()
			dep.Wait()
			if dep.Active() != unit.Active {
				u.Log.Printf("Dependency %s failed to start", name)
				err = ErrDepFail
			}
		}( /*name, unit*/ )
	}

	wg.Wait()
	if err != nil {
		return
	}

	return u.Supervisable.Start()
}
func (u *Unit) Stop() (err error) {
	if u.Supervisable == nil {
		return ErrNotLoaded
	}

	return u.Supervisable.Stop()
}
func (u *Unit) Wait() {
	if u.loading == nil {
		return
	}
	<-u.loading
	return
}

func (u *Unit) isActive() bool {
	return u.Active() == unit.Active
}
func (u *Unit) isLoading() bool {
	return u.loading == nil
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
