package unit

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

const DEFAULT_SERVICE_TYPE = "simple"

var ErrNotStarted = errors.New("Service not started")

var supportedServiceTypes = map[string]bool{
	"oneshot": true,
	"simple":  true,
	"forking": false,
	"dbus":    false,
	"notify":  false,
	"idle":    false,
}

func SupportedService(typ string) (is bool) {
	_, is = supportedServiceTypes[typ]
	return
}

// Service unit
type Service struct {
	Definition struct {
		Definition
		Service struct {
			Type                            string
			ExecStart, ExecStop, ExecReload string
			PIDFile                         string
			Restart                         string
			RemainAfterExit                 bool
		}
	}

	*exec.Cmd
}

// Define attempts to fill the sv definition by parsing r
func (sv *Service) Define(r io.Reader) (err error) {
	service := Service{}

	def := &service.Definition
	def.Service.Type = DEFAULT_SERVICE_TYPE

	if err = ParseDefinition(r, def); err != nil {
		return
	}

	merr := MultiError{}

	// Check definition for errors
	switch {
	case def.Service.ExecStart == "":
		merr = append(merr, ParseErr("ExecStart", ErrNotSet))

	case !SupportedService(def.Service.Type):
		merr = append(merr, ParseErr("Type", ParseErr(def.Service.Type, ErrNotSupported)))
	}

	if len(merr) > 0 {
		return merr
	}

	*sv = service

	return nil
}

func parseCommand(ExecStart string) *exec.Cmd {
	cmd := strings.Fields(ExecStart)
	return exec.Command(cmd[0], cmd[1:]...)
}

// Start executes the command specified in service definition
func (sv *Service) Start() (err error) {
	if sv.Cmd == nil {
		sv.Cmd = parseCommand(sv.Definition.Service.ExecStart)
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
		return ErrNotStarted
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
