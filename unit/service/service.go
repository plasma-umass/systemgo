// Package service defines a service unit type
package service

import (
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/b1101/systemgo/unit"
)

const DEFAULT_TYPE = "simple"

var ErrNotStarted = errors.New("Service not started")

var supported = map[string]bool{
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
	}
}

func Supported(typ string) (is bool) {
	return supported[typ]
}

// Define attempts to fill the sv definition by parsing r
func (sv *Unit) Define(r io.Reader) (err error) {
	def := Definition{}
	def.Service.Type = DEFAULT_TYPE

	if err = unit.ParseDefinition(r, &def); err != nil {
		return
	}

	merr := unit.MultiError{}

	// Check definition for errors
	switch {
	case def.Service.ExecStart == "":
		merr = append(merr, unit.ParseErr("ExecStart", unit.ErrNotSet))

	case !Supported(def.Service.Type):
		merr = append(merr, unit.ParseErr("Type", unit.ParseErr(def.Service.Type, unit.ErrNotSupported)))
	}

	if len(merr) > 0 {
		return merr
	}

	sv.Definition = def

	return nil
}

func parseCommand(ExecStart string) *exec.Cmd {
	cmd := strings.Fields(ExecStart)
	return exec.Command(cmd[0], cmd[1:]...)
}

// Start executes the command specified in service definition
func (sv *Unit) Start() (err error) {
	if sv.Cmd == nil {
		sv.Cmd = parseCommand(sv.Definition.Service.ExecStart)
	}
	switch sv.Definition.Service.Type {
	case "simple":
		return sv.Cmd.Start()
	case "oneshot":
		return sv.Cmd.Run()
	default:
		return unit.ErrNotSupported
	}
}

// Stop stops execution of the command specified in service definition
func (sv *Unit) Stop() (err error) {
	if sv.Cmd == nil {
		return ErrNotStarted
	}
	return sv.Process.Kill()
}

// Active reports activation status of a service
func (sv *Unit) Active() unit.Activation {
	switch {
	case sv.Cmd == nil, sv.ProcessState == nil:
		return unit.Inactive
	case sv.ProcessState.Success():
		switch sv.Definition.Service.Type {
		case "oneshot":
			return unit.Active
		case "simple":
			return unit.Inactive
		default:
			return unit.Failed
		}
	case sv.ProcessState.Exited():
		return unit.Failed
	default:
		return unit.Inactive
	}
}

// Sub reports the sub status of a service
func (sv *Unit) Sub() string {
	//return fmt.Sprint(Dead) // TODO: fix
	return "TODO: implement"
}
