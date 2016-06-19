package unit

import (
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

const testName = "foo"
const testSuffix = ".wrong"

func TestConstructorOfName(t *testing.T) {
	for suffix, _ := range constructors {
		if constr, err := ConstructorOfName(testName + suffix); err != nil {
			t.Errorf(test.Error, err)
		} else if constr == nil {
			t.Errorf(test.Nil, suffix+" constructor")
		}
	}
	_, err := ConstructorOfName(testName + testSuffix)
	if err == nil {
		t.Errorf(test.KnownType, testSuffix)
	} else if err != ErrUnknownType {
		t.Errorf(test.Mismatch, "error", err, ErrUnknownType)
	}
}

func TestSuportedName(t *testing.T) {
	for suffix, _ := range constructors {
		if !SupportedName(testName + suffix) {
			t.Errorf(test.NotSupported, suffix)
		}
	}

	if SupportedName(testName + testSuffix) {
		t.Errorf(test.Supported, testSuffix)
	}
}
