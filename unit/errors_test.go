package unit_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/b1101/systemgo/test"
	"github.com/b1101/systemgo/unit"
)

var ErrTest = errors.New("test")
var source = "test"
var expected = fmt.Sprintf("%s: %s", source, ErrTest)

func TestParseErr(t *testing.T) {
	pe := unit.ParseErr(source, ErrTest)

	if pe.Source != "test" {
		t.Errorf(test.MismatchIn, "pe.Source", pe.Source, source)
	}

	if pe.Err != ErrTest {
		t.Errorf(test.MismatchIn, "pe.Err", pe.Err, ErrTest)
	}

	if pe.Error() != expected {
		t.Errorf(test.MismatchIn, "pe.Error()", pe.Error(), expected)
	}
}

var errCount = 5

func TestMultiErr(t *testing.T) {
	me := unit.MultiError{}

	for i := 0; i < errCount; i++ {
		me = append(me, ErrTest)
	}
	if len(me) != errCount {
		t.Errorf(test.MismatchInVal, "len(me)", len(me), errCount)
	}
	for i, msg := range me.Errors() {
		if msg != me[i].Error() {
			t.Errorf(test.MismatchIn, "me.Errors()", msg, me[i].Error())
		}
	}
}
