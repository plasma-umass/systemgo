package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuported(t *testing.T) {
	for typ, is := range supported {
		assert.Equal(t, is, Supported(typ), typ)
	}

	assert.False(t, Supported("not-a-service"))
}
