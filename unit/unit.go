package unit

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
)

type Supervisable interface {
	Start()
	Stop()

	Init()

	SetPath(string)
	SetLoaded(LoadState)

	SetOutput(io.Writer)
	Log(v ...interface{})

	//Path() string
	Status() Status
	Enabled() EnableState
	Loaded() LoadState
	Active() ActivationState
	Sub() fmt.Stringer
}
type Reloader interface {
	Reload()
}

type StartStopper interface {
	Starter
	Stopper
}
type Starter interface {
	Start()
}
type Stopper interface {
	Stop()
}

// Struct representing the unit
type Unit struct {
	*log.Logger
	Buffer bytes.Buffer

	Deps      []Supervisable
	Conflicts []Supervisable

	name   string
	path   string
	loaded LoadState
	//*Definition
}

type Definition struct {
	Unit struct {
		Description                               string
		Documentation                             []string
		After, Wants, Requires, Conflicts, Before []string
	}
	Install struct {
		WantedBy string
	}
}

// Activation status of a unit
type ActivationState int

//go:generate stringer -type=ActivationState
const (
	Inactive ActivationState = iota
	Activating
	Active
	Failed // TODO: check
)

// Enable status of a unit
type EnableState int

//go:generate stringer -type=EnableState
const (
	Disabled EnableState = iota
	Static
	Indirect
	Enabled
)

// Load status of a unit
type LoadState int

//go:generate stringer -type=LoadState
const (
	Loaded LoadState = iota
	Error
)

type Status struct {
	Load LoadStatus
	Act  ActivationStatus
}
type ActivationStatus struct {
	Status ActivationState
	Sub    string
}
type LoadStatus struct {
	Status LoadState
	Path   string
	State  EnableState
	Vendor VendorStatus
}
type VendorStatus struct {
	State EnableState
}

func isUp(u Supervisable) bool {
	switch u.Active() {
	case Active:
		return true
	default:
		return false
	}
}

func (u *Unit) Log(v ...interface{}) {
	u.Logger.Println(v)
}
func (u *Unit) Read(p []byte) (n int, err error) {
	return u.Buffer.Read(p)
}

func (u *Unit) Init() {
	u.Logger = log.New(&u.Buffer, "", log.LstdFlags)
	//u.Read = u.Log.Read
}
func (u *Unit) SetLoaded(state LoadState) {
	u.loaded = state
}
func (u *Unit) SetPath(path string) {
	u.path = path
}

func (u Unit) Name() string {
	return u.name
}
func (u Unit) Path() string {
	return u.path
}
func (u Unit) Enabled() EnableState {
	return Enabled // TODO: fixme
}
func (u Unit) Loaded() LoadState {
	return u.loaded
}
func (u Unit) Active() ActivationState {
	return Active // TODO: fixme
}
func (u Unit) Sub() string {
	return "sub" // TODO: fixme
}
func (u Unit) Status() Status {
	return Status{
		LoadStatus{u.Loaded(), u.Path(), u.Enabled(), VendorStatus{u.Enabled()}}, // TODO:fixme
		ActivationStatus{u.Active(), u.Sub()},
	}
}

func (s Status) String() string {
	return fmt.Sprintf(`Loaded: %s
Active: %s`, s.Load, s.Act)
}

func (s LoadStatus) String() string {
	return fmt.Sprintf("%s (%s; %s; %s)",
		s.Status, s.Path, s.State, s.Vendor)
}
func (s VendorStatus) String() string {
	return fmt.Sprintf("vendor preset: %s",
		s.State)
}
func (s ActivationStatus) String() string {
	return fmt.Sprintf("%s (%s)",
		s.Status, s.Sub)
}

// Starts unit's dependencies
func (u *Unit) Start() {
	for _, dep := range u.Conflicts {
		if isUp(dep) {
			//u.Println("Conflicts with", dep.Path())
		}
	}

	for _, dep := range u.Deps {
		if !isUp(dep) {
			dep.Start()
		}
	}

	for _, dep := range u.Deps {
		for !isUp(dep) {
			time.Sleep(300 * time.Millisecond)
		}
	}
}

//for _, name := range u.Unit.Conflicts {
//if dep, ok := u.GS.GetUnit(name); ok {
//if isUp(dep) {
//return errors.New("conflicts with " + name)
//}
//}
//}

//for _, name := range u.Unit.After {
//if dep, ok := u.GS.GetUnit(name); !ok {
//return errors.New(name + " not found")
//} else {
//for !isUp(dep) {
//log.Println("waiting for", name)
//time.Sleep(time.Second)
//}
//}
//}
//// Error checking, redundant
//for _, name := range u.Unit.Before {
//if dep, ok := u.GS.GetUnit(name); !ok {
//return errors.New(name + " not found")
//} else {
//if isUp(dep) {
//return errors.New(name + " already started")
//}
//}
//}

//for _, name := range u.Unit.Requires {
//go func() {
//if dep, ok := u.GS.GetUnit(name); !ok {
//Errs <- errors.New(name + " not found")
//} else {
//if dep.Status == Loading {
//return
//}
//if !isUp(dep) {
//log.Println("starting", name)
//if err = dep.Start(); err != nil {
//Errs <- errors.New("Error starting " + name + ": " + err.Error())
//return
//}
//if dep.GetStatus() != Active {
//Errs <- errors.New(name + " failed to launch")
//}
//}
//}
//}()
//}
//for _, name := range u.Unit.Wants {
//go func() {
//if dep, ok := u.GS.GetUnit(name); !ok {
//Errs <- errors.New(name + " not found")
//} else {
//if !isUp(dep) {
//log.Println("starting", name)
//if err = dep.Start(); err != nil {
//Errs <- errors.New("Error starting " + name + ": " + err.Error())
//}
//}
//}
//}()
//}
//}

// Stop and restart unit execution
//func (u *Unit) Restart() (err error) {
//if err = u.Stop(); err != nil {
//return
//}
////delete(u.Loaded, u)

//var cmd []string
//if u.Service.ExecReload == "" {
//cmd = strings.Split(u.Service.ExecStart, " ")
//} else {
//cmd = strings.Split(u.Service.ExecReload, " ")
//}
//u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

//err = u.Start()
////u.Loaded[u] = true
//return
//}

// Reload unit definition
//func (u *Unit) Reload() error {
////u., _ = ParseUnit(u)
//return errors.New("not implemented yet") // TODO
//}
