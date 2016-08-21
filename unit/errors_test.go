package unit_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/plasma-umass/systemgo/unit"
	"github.com/stretchr/testify/assert"
)

var ErrTest = errors.New("test")
var source = "test"
var expected = fmt.Sprintf("%s: %s", source, ErrTest)

func TestParseErr(t *testing.T) {
	pe := unit.ParseErr(source, ErrTest)

	assert.Equal(t, pe.Source, "test", "pe.Source")
	assert.Equal(t, pe.Err, ErrTest, "pe.Err")
	assert.EqualError(t, pe, expected)
}

var errCount = 5

func TestMultiErr(t *testing.T) {
	me := unit.MultiError{}

	for i := 0; i < errCount; i++ {
		me = append(me, ErrTest)
	}

	assert.Len(t, me, errCount)

	for i, msg := range me.Errors() {
		assert.EqualError(t, me[i], msg)
	}
}
