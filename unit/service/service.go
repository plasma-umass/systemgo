package service

import (
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	*unit.Unit
	*Definition
	*exec.Cmd
}
type Definition struct {
	*unit.Definition
	Service struct {
		Type                            string
		ExecStart, ExecStop, ExecReload string
		PIDFile                         string
		Restart                         string
		RemainAfterExit                 bool
		//Type                        ServiceType
	}
}

func New(definition io.Reader) (*Unit, error) {
	service := &Unit{}
	service.Unit = unit.New()
	service.Definition = &Definition{Definition: service.Unit.Definition}

	if err := unit.Define(definition, service.Definition); err != nil {
		return nil, err
	}

	switch def := service.Definition; {
	case def.Service.ExecStart == "":
		return nil, errors.New("ExecStart field not set")
		fallthrough
	case def.Service.Type == "":
		def.Service.Type = "simple"
		fallthrough
	default:
		cmd := strings.Split(def.Service.ExecStart, " ")
		service.Cmd = exec.Command(cmd[0], cmd[1:]...)

		return service, nil
	}
}

func (u *Unit) Start() (err error) {
	switch u.Service.Type {
	case "simple":
		err = u.Cmd.Start()
	case "oneshot":
		err = u.Cmd.Run()
	case "forking", "dbus", "notify", "idle":
		err = errors.New(u.Service.Type + " not implemented yet") // TODO
	default:
		err = errors.New(u.Service.Type + " does not exist")
	}
	return
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() (err error) {
	return u.Process.Kill()
}

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

func (u Unit) Sub() string {
	//return fmt.Sprint(Dead) // TODO: fix
	return "TODO: implement"
}
