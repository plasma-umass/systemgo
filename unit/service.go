// Package service defines a service unit type
package unit

import (
	"io"
	"os/exec"
	"strings"
)

const DEFAULT_SERVICE_TYPE = "simple"

var supportedServiceTypes = map[string]bool{
	"oneshot": true,
	"simple":  true,
	"forking": false,
	"dbus":    false,
	"notify":  false,
	"idle":    false,
}

func SupportedServiceType(typ string) (is bool) {
	_, is = supportedServiceTypes[typ]
	return
}

// Service unit
type Service struct {
	Definition serviceDefinition
	*exec.Cmd
}

// Service unit definition
type serviceDefinition struct {
	Definition
	Service struct {
		Type                            string
		ExecStart, ExecStop, ExecReload string
		PIDFile                         string
		Restart                         string
		RemainAfterExit                 bool
	}
}

// Define attempts to fill the sv definition by parsing r
func (sv *Service) Define(r io.Reader) (err error) {
	def := serviceDefinition{}

	def.Service.Type = DEFAULT_SERVICE_TYPE

	if err = ParseDefinition(r, &def); err != nil {
		return
	}

	// Check Definition for errors
	switch {
	case def.Service.ExecStart == "":
		return ParseErr("ExecStart", ErrNotSet)

	case !SupportedServiceType(def.Service.Type):
		return ParseErr("Type", ParseErr(def.Service.Type, ErrNotSupported))
	}

	// Only load definition of a unit if it is correct (possibly overwriting existing one)
	sv.Definition = def

	cmd := strings.Fields(def.Service.ExecStart)
	sv.Cmd = exec.Command(cmd[0], cmd[1:]...)

	return
}

// Start executes the command specified in service definition
func (sv *Service) Start() (err error) {
	if sv.Cmd == nil {
		return ErrNotLoaded
	}
	switch sv.Definition.Service.Type {
	case "simple":
		return sv.Cmd.Start()
	case "oneshot":
		return sv.Cmd.Run()
	default:
		return ErrNotSupported
	}
}

// Stop stops execution of the command specified in service definition
func (sv *Service) Stop() (err error) {
	if sv.Cmd == nil {
		return ErrNotLoaded
	}
	return sv.Process.Kill()
}

// Active reports activation status of a service
func (sv Service) Active() Activation {
	switch {
	case sv.Cmd == nil, sv.ProcessState == nil:
		return Inactive
	case sv.ProcessState.Success():
		switch sv.Definition.Service.Type {
		case "oneshot":
			return Active
		case "simple":
			return Inactive
		default:
			return Failed
		}
	case sv.ProcessState.Exited():
		return Failed
	default:
		return Inactive
	}
}

// Sub reports the sub status of a service
func (sv Service) Sub() string {
	//return fmt.Sprint(Dead) // TODO: fix
	return "TODO: implement"
}
