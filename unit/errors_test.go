package unit

import (
	"errors"
	"fmt"
	"testing"

	"github.com/b1101/systemgo/test"
)

var ErrTest = errors.New("test")
var source = "test"
var expected = fmt.Sprintf("%s: %s", source, ErrTest)

func TestParseErr(t *testing.T) {
	pe := ParseErr(source, ErrTest)

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
