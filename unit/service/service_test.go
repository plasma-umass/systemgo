package service_test

import (
	"strings"
	"testing"

	"github.com/b1101/systemgo/unit"
	"github.com/b1101/systemgo/unit/service"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	sv := service.Unit{}

	err := sv.Define(strings.NewReader(`[Service]
ExecStart=/bin/echo test`))
	assert.NoError(t, err, "sv.Define")
	assert.Equal(t, sv.Definition.Service.Type, service.DEFAULT_TYPE, "sv.Definition.Service.Type")

	sv = service.Unit{}

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
func TestSimpleService(t *testing.T) {
	sv := service.Unit{}

	contents := strings.NewReader(`[Service]
ExecStart=/bin/sleep 5
Type=simple`)

	assert.NoError(t, sv.Define(contents), "sv.Define")
	assert.NoError(t, sv.Start(), "sv.Start")
	assert.NotNil(t, sv.Cmd.Process)
}

// Oneshot service type test
func TestOneshotService(t *testing.T) {
	sv := service.Unit{}

	contents := strings.NewReader(`[Service]
ExecStart=/bin/echo oneshot
Type=oneshot`)

	assert.NoError(t, sv.Define(contents), "sv.Define")
	assert.NoError(t, sv.Start(), "sv.Start")
	if assert.NotNil(t, sv.Cmd.ProcessState, "sv.Cmd.ProcessState") {
		assert.True(t, sv.Cmd.ProcessState.Success())
	}
}
