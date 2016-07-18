package service

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/b1101/systemgo/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefine(t *testing.T) {
	sv := Unit{}

	err := sv.Define(strings.NewReader(`[Service]
ExecStart=/bin/echo test`))
	assert.NoError(t, err, "sv.Define")
	assert.Equal(t, sv.Definition.Service.Type, DEFAULT_TYPE, "sv.Definition.Service.Type")

	sv = Unit{}

	if err = sv.Define(strings.NewReader(`[Service]`)); assert.Error(t, err, "sv.Define with wrong definition") {
		if me, ok := err.(unit.MultiError); assert.True(t, ok, "error is MultiError") {
			if pe, ok := me[0].(unit.ParseError); assert.True(t, ok, "error is ParseError") {
				assert.Equal(t, "ExecStart", pe.Source)
				assert.Equal(t, unit.ErrNotSet, pe.Err)
			}
		}
	}
}

// Simple service type test
func TestStartSimple(t *testing.T) {
	sv := Unit{}
	sv.Definition.Service.Type = "simple"
	sv.Definition.Service.ExecStart = "/bin/sleep 60"

	require.NoError(t, sv.Start(), "sv.Start")
	assert.NotNil(t, sv.Cmd.Process)
	assert.Nil(t, sv.Cmd.ProcessState)
}

func TestStartOneshot(t *testing.T) {
	sv := Unit{}
	sv.Definition.Service.Type = "oneshot"
	sv.Definition.Service.ExecStart = "/bin/echo test"

	assert.NoError(t, sv.Start())
	assert.NotNil(t, sv.Cmd.Process)
	if assert.NotNil(t, sv.Cmd.ProcessState) {
		assert.True(t, sv.Cmd.ProcessState.Success())
	}
}

func TestActive(t *testing.T) {
	// Oneshot service
	sv := Unit{}
	sv.Cmd = exec.Command("echo", "test")

	sv.Definition.Service.Type = "oneshot"
	sv.Definition.Service.RemainAfterExit = true
	if assert.NoError(t, sv.Cmd.Run(), "oneshot Cmd.Run()") {
		assert.Equal(t, unit.Active, sv.Active(), fmt.Sprintf("oneshot service - %s", sv.Active()))
	}

	// Simple service
	sv = Unit{}
	sv.Cmd = exec.Command("sleep", "60")
	sv.Definition.Service.Type = "simple"
	if assert.NoError(t, sv.Cmd.Start(), "simple Cmd.Run()") {
		assert.Equal(t, unit.Active, sv.Active(), fmt.Sprintf("simple service - %s", sv.Active()))
	}
	// TODO
	//assert.NoError(t, sv.Cmd.Process.Kill())
	//assert.Equal(t, unit.Failed, sv.Active(), fmt.Sprintf("Active(): %s, Sub(): %s", sv.Active(), sv.sub()))

	sv = Unit{}
	sv.starting = true
	assert.Equal(t, unit.Activating, sv.Active())

	sv = Unit{}
	sv.reloading = true
	assert.Equal(t, unit.Reloading, sv.Active())

	sv = Unit{}
	assert.Equal(t, unit.Inactive, sv.Active())

}

func TestSuported(t *testing.T) {
	for typ, is := range supported {
		assert.Equal(t, is, Supported(typ), typ)
	}

	assert.False(t, Supported("not-a-service"))
}
