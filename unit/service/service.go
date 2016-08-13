// Package service defines a service unit type
package service

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/rvolosatovs/systemgo/unit"
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

	starting, reloading bool
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
	sv.starting = true
	defer func() {
		sv.starting = false
	}()

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

func (sv *Unit) IsReloading() bool {
	return sv.reloading
}

func (sv *Unit) IsStarting() bool {
	return sv.starting
}

// Active reports activation status of a service
func (sv *Unit) Active() unit.Activation {
	// based of Systemd transtition table found in https://goo.gl/oEjikJ
	switch sv.sub() {
	case Dead:
		return unit.Inactive
	case Failed:
		return unit.Failed
	case Reload:
		return unit.Reloading
	case Running, Exited:
		return unit.Active
	case Start, StartPre, StartPost, AutoRestart:
		return unit.Activating
	case Stop, StopSigabrt, StopPost, StopSigkill, StopSigterm, FinalSigkill, FinalSigterm:
		return unit.Deactivating
	default:
		// Unreachable
		return -1
	}
}

// Sub reports the sub status of a service
func (sv *Unit) Sub() string {
	return sv.sub().String()
}

func (sv *Unit) sub() (s Sub) {
	switch def := sv.Definition.Service; {
	case sv.IsReloading():
		return Reload

	case sv.IsStarting():
		return Start

	case sv.Cmd == nil, sv.Cmd.Process == nil:
		// Service has not been started yet
		return Dead

	case sv.Cmd.ProcessState == nil:
		return Running
		// TODO: find a way to distinguish between Failed and Running processes
		//if isRunning(sv.Cmd) {
		//return Running
		//} else if sv.Cmd.ProcessState == nil {
		//return Failed
		//}
		//fallthrough

	case sv.ProcessState.Success() && def.Type == "oneshot" && def.RemainAfterExit:
		// Service process successfully exited
		return Exited

	case sv.ProcessState.Exited():
		return Stop

		// TODO (if needed)
		// // Should be safe on Unix
		// switch st := sv.ProcessState.Sys().(syscall.WaitStatus); {
		// }

	default:
		// Service process has finished, but did not return a 0 exit code
		return Failed
	}
}

// Does not work
func isRunning(cmd *exec.Cmd) (running bool) {
	running = true

	defer func() {
		fmt.Println("recovering")
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	go func() {
		time.Sleep(time.Millisecond)
		if running {
			panic("")
		}
	}()

	cmd.Process.Wait()
	running = false
	return
}
