package service_test

import (
	"strings"
	"testing"

	"github.com/b1101/systemgo/test"
	"github.com/b1101/systemgo/unit"
	"github.com/b1101/systemgo/unit/service"
)

func TestSupportedSimple(t *testing.T) {
	if !service.Supported("simple") {
		t.Errorf(test.MismatchVal, false, true)
	}
}

func TestNewService(t *testing.T) {
	sv := service.Unit{}

	err := sv.Define(strings.NewReader(`[Service]
ExecStart=/bin/echo test`))
	if err != nil {
		t.Errorf(test.ErrorIn, "sv.Define", err)
	}

	if sv.Service.Type != service.DEFAULT_TYPE {
		t.Errorf(test.MismatchIn, "sv.Definition.Service.Type", sv.Service.Type, service.DEFAULT_TYPE)
	}

	sv = service.Unit{}

	if err = sv.Define(strings.NewReader(`[Service]`)); err != nil {
		if me, ok := err.(unit.MultiError); ok {
			if pe, ok := me[0].(unit.ParseError); ok {
				if pe.Source == "ExecStart" && pe.Err == unit.ErrNotSet {
					return
				}
				t.Errorf(test.MismatchIn, "pe.Err", pe.Err, unit.ErrNotSet)
			}
			t.Errorf(test.MismatchInType, "err", err, unit.ParseError{})
		}
		t.Errorf(test.MismatchInType, "err", err, unit.MultiError{})
	}
	t.Errorf(test.NotDetected, "empty ExecStart field")
}

// Simple service type test
func TestSimpleService(t *testing.T) {
	sv := service.Unit{}

	sv.Definition.Service.ExecStart = "/bin/sleep 5"

	err := sv.Define(strings.NewReader(`[Service]
ExecStart=/bin/sleep 5
Type=simple`))
	if err != nil {
		t.Errorf(test.ErrorIn, "sv.Define", err)
	}

	if err = sv.Start(); err != nil {
		t.Errorf(test.ErrorIn, "sv.Start()", err)
	}

	if process := sv.Cmd.Process; process == nil {
		t.Errorf(test.Nil, "sv.Cmd.Process")
	}
}

// Oneshot service type test
func TestOneshotService(t *testing.T) {
	sv := service.Unit{}

	err := sv.Define(strings.NewReader(`[Service]
ExecStart=/bin/echo oneshot
Type=oneshot`))
	if err != nil {
		t.Errorf(test.ErrorIn, "unit.NewService", err)
	}

	if err = sv.Start(); err != nil {
		t.Errorf(test.ErrorIn, "sv.Start()", err)
	}

	if state := sv.Cmd.ProcessState; state != nil {
		if !state.Success() {
			t.Errorf("Process exited with failure, pid: %v", state.Pid())
		}
	} else {
		//t.Errorf("Process state is nil.\n process: %v", sv.Cmd.Process)
		t.Errorf(test.Nil, "sv.Cmd.ProcessState")
	}
}
