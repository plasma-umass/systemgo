package service

import (
	"strings"
	"testing"

	"github.com/b1101/systemgo/unit"
)

func TestNew(t *testing.T) {
	sv, err := New(strings.NewReader(`[Service]
ExecStart=/bin/echo test`))
	if err != nil {
		t.Errorf("Error creating unit: %s", err)
	}

	if sv.Service.Type != DEFAULT_SERVICE_TYPE {
		t.Errorf("Default service type not set: %s != %s", sv.Service.Type, DEFAULT_SERVICE_TYPE)
	}

	func() {
		if _, err = New(strings.NewReader(`[Service]`)); err != nil {
			if pe, ok := err.(unit.ParseError); ok {
				if pe.Source == "ExecStart" && pe.Err == unit.ErrNotSet {
					return
				}
				t.Errorf("Wrong ParseError received: %s", pe)
			}
			t.Errorf("Error is not ParseError: %s", err)
		}
		t.Errorf("Empty ExecStart field not detected")
	}()
}

func TestStart(t *testing.T) {
	// Simple service type test
	sv, err := New(strings.NewReader(`[Service]
ExecStart=/bin/sleep 5
Type=simple`))
	if err != nil {
		t.Errorf("Error creating unit: %s", err)
	}

	if err = sv.Start(); err != nil {
		t.Errorf("Error starting unit: %s", err)
	}

	if process := sv.Cmd.Process; process == nil {
		t.Errorf("Process is nil")
	}

	// Oneshot service type test
	sv, err = New(strings.NewReader(`[Service]
ExecStart=/bin/echo oneshot
Type=oneshot`))
	if err != nil {
		t.Errorf("Error creating unit: %s", err)
	}

	if err = sv.Start(); err != nil {
		t.Errorf("Error starting unit: %s", err)
	}

	if state := sv.Cmd.ProcessState; state != nil {
		if !state.Success() {
			t.Errorf("Process exited with failure, pid: %v", state.Pid())
		}
	} else {
		t.Errorf("Process state is nil.\n process: %v", sv.Cmd.Process)
	}
}
