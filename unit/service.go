// Package service defines a service unit type
package unit

import (
	"io"
	"os/exec"
	"strings"
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
type Service struct {
	serviceDefinition
	*exec.Cmd
}

// Service unit definition
type serviceDefinition struct {
	definition
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
func NewService(definition io.Reader) (service *Service, err error) {
	service = &Service{}

	// Set defaults
	service.serviceDefinition.Service.Type = "simple"

	if err = parseDefinition(definition, &service.serviceDefinition); err != nil {
		return
	}

	// Check definition for errors
	switch def := service.serviceDefinition; {
	case def.Service.ExecStart == "":
		err = ParseErr("ExecStart", ErrNotSet)
	case !supportedTypes[def.Service.Type]:
		var terr error
		if _, ok := supportedTypes[def.Service.Type]; !ok {
			terr = ErrNotExist
		} else {
			terr = ErrNotSupported
		}
		err = ParseErr("Type", ParseErr(def.Service.Type, terr))
	default:
		cmd := strings.Fields(def.Service.ExecStart)
		service.Cmd = exec.Command(cmd[0], cmd[1:]...)
	}
	return
}

// Start executes the command specified in service definition
func (u *Service) Start() (err error) {
	switch u.Service.Type {
	case "simple":
		err = u.Cmd.Start()
	case "oneshot":
		err = u.Cmd.Run()
	default:
		err = ErrNotSupported
	}
	return
}

// Stop stops execution of the command specified in service definition
func (u *Service) Stop() (err error) {
	return u.Process.Kill()
}

// Active reports activation status of a service
func (u Service) Active() Activation {
	switch {
	case u.Cmd == nil, u.ProcessState == nil:
		return Inactive
	case u.ProcessState.Success():
		switch u.Service.Type {
		case "oneshot":
			return Active
		case "simple":
			return Inactive
		default:
			return Failed
		}
	case u.ProcessState.Exited():
		return Failed
	default:
		return Inactive
	}
}

// Sub reports the sub status of a service
func (u Service) Sub() string {
	//return fmt.Sprint(Dead) // TODO: fix
	return "TODO: implement"
}
