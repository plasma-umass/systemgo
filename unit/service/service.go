// Package service defines a service unit type
package service

import (
	"io"
	"os/exec"
	"strings"

	"github.com/plasma-umass/systemgo/unit"

	log "github.com/Sirupsen/logrus"
)

const DEFAULT_TYPE = "simple"

const (
	dead         = "dead"
	startPre     = "startPre"
	start        = "start"
	startPost    = "startPost"
	running      = "running"
	exited       = "exited" // not running anymore, but RemainAfterExit true for this unit
	reload       = "reload"
	stop         = "stop"
	stopSigabrt  = "stopSigabrt" // watchdog timeout
	stopSigterm  = "stopSigterm"
	stopSigkill  = "stopSigkill"
	stopPost     = "stopPost"
	finalSigterm = "finalSigterm"
	finalSigkill = "finalSigkill"
	failed       = "failed"
	autoRestart  = "autoRestart"
)

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
		//Restart                         string
		//RestartSec                      int
		RemainAfterExit  bool
		WorkingDirectory string
		//PIDFile          string
	}
}

func Supported(typ string) (is bool) {
	return supported[typ]
}

//func (sv *Unit) String() string {
//return sv.Service.ExecStart
//}

// Define attempts to fill the sv definition by parsing r
func (sv *Unit) Define(r io.Reader /*, errch chan<- error*/) (err error) {
	log.WithField("r", r).Debugf("sv.Define")

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

	cmd := strings.Fields(def.Service.ExecStart)
	sv.Cmd = exec.Command(cmd[0], cmd[1:]...)
	sv.Cmd.Dir = sv.Definition.Service.WorkingDirectory

	return nil
}

// Start executes the command specified in service definition
func (sv *Unit) Start() (err error) {
	e := log.WithField("ExecStart", sv.Definition.Service.ExecStart)

	e.Debug("sv.Start")

	switch sv.Definition.Service.Type {
	case "simple":
		if err = sv.Cmd.Start(); err == nil {
			go sv.Cmd.Wait()
		}
	case "oneshot":
		err = sv.Cmd.Run()
	default:
		panic("Unknown service type")
	}

	e.WithField("err", err).Debug("started")
	return
}

// Stop stops execution of the command specified in service definition
func (sv *Unit) Stop() (err error) {
	if cmd := strings.Fields(sv.Definition.Service.ExecStop); len(cmd) > 0 {
		return exec.Command(cmd[0], cmd[1:]...).Run()
	}
	if sv.Cmd.Process != nil {
		return sv.Cmd.Process.Kill()
	}
	return nil
}

// Sub reports the sub status of a service
func (sv *Unit) Sub() string {
	log.WithField("sv", sv).Debugf("sv.Sub")

	switch {
	case sv.Cmd.Process == nil:
		// Service has not been started yet
		return dead

	case sv.Cmd.ProcessState == nil:
		// Wait has not returned yet
		return running

	case sv.ProcessState.Exited(), sv.ProcessState.Success():
		if sv.Definition.Service.RemainAfterExit {
			return exited
		}
		return dead

	default:
		// Service process has finished, but did not return a 0 exit code
		return failed
	}
}

// Active reports activation status of a service
func (sv *Unit) Active() unit.Activation {
	log.WithField("sv", sv).Debugf("sv.Active")

	// based of Systemd transtition table found in https://goo.gl/oEjikJ
	switch sv.Sub() {
	case dead:
		return unit.Inactive
	case failed:
		return unit.Failed
	case reload:
		return unit.Reloading
	case running, exited:
		return unit.Active
	case start, startPre, startPost, autoRestart:
		return unit.Activating
	case stop, stopSigabrt, stopPost, stopSigkill, stopSigterm, finalSigkill, finalSigterm:
		return unit.Deactivating
	default:
		panic("Unknown service sub state")
	}
}
