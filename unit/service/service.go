package service

import (
	"errors"
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

func New() unit.Supervisable {
	service := &Unit{}
	service.Unit = unit.New()
	service.Definition = &Definition{Definition: service.Unit.Definition}
	return service
}

func (u *Unit) Start() {
	var err error

	u.Unit.Start()

	cmd := strings.Split(u.Service.ExecStart, " ")
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

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

	if err != nil {
		u.Log(err.Error())
	}
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() {
	if err := u.Process.Kill(); err != nil {
		u.Log(err.Error())
	}
}

func (u *Unit) Active() unit.ActivationState {
	return unit.Active // TODO: fixme
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
		return unit.Active
	}
}
