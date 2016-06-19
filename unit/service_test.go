package unit

import (
	"strings"
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

func TestNewService(t *testing.T) {
	sv, err := NewService(strings.NewReader(`[Service]
ExecStart=/bin/echo test`))
	if err != nil {
		t.Errorf(test.ErrorIn, "NewService", err)
	}

	if sv.Service.Type != DEFAULT_SERVICE_TYPE {
		t.Errorf(test.MismatchIn, "sv.Service.Type", sv.Service.Type, DEFAULT_SERVICE_TYPE)
	}

	func() {
		if _, err = NewService(strings.NewReader(`[Service]`)); err != nil {
			if pe, ok := err.(ParseError); ok {
				if pe.Source == "ExecStart" && pe.Err == ErrNotSet {
					return
				}
				t.Errorf(test.MismatchIn, "pe.Err", pe.Err, ErrNotSet)
			}
			t.Errorf(test.MismatchInType, "err", err, ParseError{})
		}
		t.Errorf(test.NotDetected, "empty ExecStart field")
	}()
}

// Simple service type test
func TestSimple(t *testing.T) {
	sv, err := NewService(strings.NewReader(`[Service]
ExecStart=/bin/sleep 5
Type=simple`))
	if err != nil {
		t.Errorf(test.ErrorIn, "NewService", err)
	}

	if err = sv.Start(); err != nil {
		t.Errorf(test.ErrorIn, "sv.Start()", err)
	}

	if process := sv.Cmd.Process; process == nil {
		t.Errorf(test.Nil, "sv.Cmd.Process")
	}
}

// Oneshot service type test
func TestOneshot(t *testing.T) {
	sv, err := NewService(strings.NewReader(`[Service]
ExecStart=/bin/echo oneshot
Type=oneshot`))
	if err != nil {
		t.Errorf(test.ErrorIn, "NewService", err)
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
