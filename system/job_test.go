package system

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobState(t *testing.T) {
	j := newJob(-1, nil)
	assert.Equal(t, running, j.State())

	j.finish()
	assert.Equal(t, success, j.State())
	assert.True(t, j.Success())

	j.err = errors.New("")
	assert.Equal(t, failed, j.State())
}
