package unit

import (
	"errors"
	"log"
	"os/exec"
	"strings"
	"time"
)

var (
	Units  map[string]*Unit
	Loaded map[*Unit]bool
)

var Errs = make(chan error, 1)

// Struct representing the unit
type Unit struct {
	*Definition
	*exec.Cmd
	Status status
	State  state
	//Name      string
	//ShortName string
}

// Definition as found in the unit specification file
type Definition struct {
	Unit struct {
		Description                               string
		After, Wants, Requires, Conflicts, Before []string
	}
	Service struct {
		//Type                        ServiceType
		Type, ExecStart, ExecReload, WorkingDirectory string
	}
	Install struct {
		// WIP
	}
}

//type ServiceType string

//const (
//Simple  ServiceType = ""
//Oneshot             = "oneshot"
//Forking             = "forking"
//)

// Execution status of a unit
type status int

const (
	Inactive status = iota
	Loading
	Active
	Exited
	Failed
)

type state int

const (
	Static state = iota
	Indirect
	Disabled
	Enabled
)

// Starts execution of the unit's specified command
func (u *Unit) Start() (err error) {
	if u.Status == Loading {
		return
	} else {
		u.Status = Loading
	}

	cmd := strings.Split(u.Service.ExecStart, " ")
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	for _, name := range u.Definition.Unit.Conflicts {
		if dep, ok := Units[name]; ok {
			if Loaded[dep] {
				return errors.New("conflicts with " + name)
			}
		}
	}
	for _, name := range u.Definition.Unit.Requires {
		go func() {
			if dep, ok := Units[name]; !ok {
				Errs <- errors.New(name + " not found")
			} else {
				if dep.Status == Loading {
					return
				}
				if !Loaded[dep] {
					log.Println("starting", name)
					if err = dep.Start(); err != nil {
						Errs <- errors.New("Error starting " + name + ": " + err.Error())
						return
					}
					if dep.GetStatus() != Active {
						Errs <- errors.New(name + " failed to launch")
					}
				}
			}
		}()
	}
	for _, name := range u.Definition.Unit.Wants {
		go func() {
			if dep, ok := Units[name]; !ok {
				Errs <- errors.New(name + " not found")
			} else {
				if !Loaded[dep] {
					log.Println("starting", name)
					if err = dep.Start(); err != nil {
						Errs <- errors.New("Error starting " + name + ": " + err.Error())
					}
				}
			}
		}()
	}
	for _, name := range u.Definition.Unit.After {
		if dep, ok := Units[name]; !ok {
			return errors.New(name + " not found")
		} else {
			for !Loaded[dep] {
				log.Println("waiting for", name)
				time.Sleep(time.Second)
			}
		}
	}

	switch u.Service.Type {
	case "simple":
		err = u.Cmd.Start()
	case "oneshot":
		err = u.Cmd.Run()
	case "forking", "dbus", "notify", "idle":
		return errors.New(u.Service.Type + " not implemented yet") // TODO
	default:
		return errors.New(u.Service.Type + " does not exist")
	}
	Loaded[u] = true

	return
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() (err error) {
	if !Loaded[u] {
		return errors.New("unit not loaded")
	}

	err = u.Process.Kill()
	delete(Loaded, u)

	return
}

// Stop and restart unit execution
func (u *Unit) Restart() (err error) {
	if err = u.Stop(); err != nil {
		return
	}
	delete(Loaded, u)

	var cmd []string
	if u.Service.ExecReload == "" {
		cmd = strings.Split(u.Service.ExecStart, " ")
	} else {
		cmd = strings.Split(u.Service.ExecReload, " ")
	}
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	err = u.Cmd.Start()
	Loaded[u] = true
	return
}

// Reload unit definition
func (u *Unit) Reload() error {
	//u.Definition, _ = ParseUnit(u)
	return errors.New("not implemented yet") // TODO
}

func (u *Unit) GetStatus() status {
	// TODO: Fixme, check proper state
	switch {
	case u.Cmd.ProcessState == nil:
		switch u.Status {
		case Failed:
			return Failed
		default:
			return Inactive
		}
	case u.Cmd.ProcessState.Success():
		switch u.Service.Type {
		case "oneshot":
			return Active
		case "simple":
			return Exited
		default:
			return Failed
		}
	case u.Cmd.ProcessState.Exited():
		return Failed
	default:
		return Active
	}
}
