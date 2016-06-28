package system

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/lib/test"
	"github.com/b1101/systemgo/unit"
)

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

		if unit.SupportedName(name) {
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
