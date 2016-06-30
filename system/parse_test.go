package system

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/test"
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
		t.Errorf(test.MismatchIn, "err", err, ErrUnknownType)
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

func TestPathset(t *testing.T) {
	path, err := ioutil.TempDir("", "pathset-test")
	if err != nil {
		t.Fatalf("Error creating dir: %s", err)
	}
	defer os.RemoveAll(path)

	cases := []string{
		"test.service",
		"test.mount",
		"test.socket",
		"test.target",
		"test.wrong",
	}

	correct := 0
	for _, name := range cases {
		if err = ioutil.WriteFile(filepath.Join(path, name), []byte{}, 0666); err != nil {
			t.Fatalf(test.ErrorIn, "ioutil.WriteFile", err)
		}

		if SupportedName(name) {
			correct++
		}
	}

	if paths, err := pathset(path); err != nil {
		t.Fatalf(test.ErrorIn, "pathset", err)
	} else if len(paths) != correct {
		t.Errorf(test.MismatchInVal, "len(paths)", len(paths), correct)
	}
}

func TestParse(t *testing.T) {
	path, err := ioutil.TempDir("", "parse-test")
	if err != nil {
		t.Fatalf(test.ErrorIn, "ioutil.TempDir", err)
	}
	defer os.RemoveAll(path)

	contents := []byte(`[Service]
ExecStart=/bin/sleep 60
`)

	path = filepath.Join(path, "test.service")
	if err = ioutil.WriteFile(path, contents, 0666); err != nil {
		t.Errorf(test.ErrorIn, "ioutil.WriteFile", err)
	}

	if _, err := parseFile(path); err != nil {
		t.Errorf(test.ErrorIn, "parseFile", err)
	}
}
