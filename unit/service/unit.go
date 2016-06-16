// Package service defines a service unit type
package service

import (
	"io"
	"os/exec"
	"strings"

	"github.com/b1101/systemgo/unit"
)

const DEFAULT_SERVICE_TYPE = "simple"

var supportedTypes = map[string]bool{
	"oneshot": true,
	"simple":  true,
	"forking": false,
	"dbus":    false,
	"notify":  false,
	"idle":    false,
}

// Service unit
type Unit struct {
	Definition
	*exec.Cmd
}

// Service unit definition
type Definition struct {
	unit.Definition
	Service struct {
		Type                            string
		ExecStart, ExecStop, ExecReload string
		PIDFile                         string
		Restart                         string
		RemainAfterExit                 bool
		//Type                        ServiceType
	}
}

// Parses the definition, checks for errors and returns a new service
func New(definition io.Reader) (service *Unit, err error) {
	service = &Unit{}

	// Set defaults
	service.Definition.Service.Type = "simple"

	if err = unit.ParseDefinition(definition, &service.Definition); err != nil {
		return
	}

	// Check definition for errors
	switch def := service.Definition; {
	case def.Service.ExecStart == "":
		err = unit.ParseErr("ExecStart", unit.ErrNotSet)
	case !supportedTypes[def.Service.Type]:
		var terr error
		if _, ok := supportedTypes[def.Service.Type]; !ok {
			terr = unit.ErrNotExist
		} else {
			terr = unit.ErrNotSupported
		}
		err = unit.ParseErr("Type", unit.ParseErr(def.Service.Type, terr))
	default:
		cmd := strings.Split(def.Service.ExecStart, " ")
		service.Cmd = exec.Command(cmd[0], cmd[1:]...)
	}
	return
}

// Start executes the command specified in service definition
func (u *Unit) Start() (err error) {
	switch u.Service.Type {
	case "simple":
		err = u.Cmd.Start()
	case "oneshot":
		err = u.Cmd.Run()
	default:
		err = unit.ErrNotSupported
	}
	return
}

// Stop stops execution of the command specified in service definition
func (u *Unit) Stop() (err error) {
	return u.Process.Kill()
}

// Active reports activation status of a service
func (u Unit) Active() unit.Activation {
	switch {
	case u.Cmd == nil, u.ProcessState == nil:
		return unit.Inactive
	case u.ProcessState.Success():
		switch u.Service.Type {
		case "oneshot":
			return unit.Active
		case "simple":
			return unit.Inactive
		default:
			return unit.Failed
		}
	case u.ProcessState.Exited():
		return unit.Failed
	default:
		return unit.Inactive
	}
}

// Sub reports the sub status of a service
func (u Unit) Sub() string {
	//return fmt.Sprint(Dead) // TODO: fix
	return "TODO: implement"
}
