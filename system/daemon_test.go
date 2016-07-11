package system

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/test"
	"github.com/b1101/systemgo/test/mock_unit"
	"github.com/golang/mock/gomock"
)

func TestLoad(t *testing.T) {
	sys := New()
	sys.SetPaths(os.TempDir())

	name := "foo.service"

	fpath := filepath.Join(os.TempDir(), name)

	file, err := os.Create(fpath)
	if err != nil {
		t.Fatalf(test.ErrorIn, "os.Create("+fpath+")", err)
	}
	defer file.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uInt := mock_unit.NewMockInterface(ctrl)
	uInt.EXPECT().Define(gomock.Any()).Return(nil).Times(2)

	u := NewUnit(uInt)

	sys.parsed[name] = u
	sys.parsed[fpath] = u

	for _, name := range []string{name, fpath} {
		if ptr, err := sys.Get(name); err != nil {
			t.Errorf(test.ErrorIn, "sys.Get("+name+")", err)
		} else {
			if u != ptr {
				t.Errorf(test.MismatchInVal, "sys.Get("+name+")", u, ptr)
			}
		}
	}
}

func TestSuported(t *testing.T) {
	for suffix, is := range supported {
		if Supported("foo"+suffix) != is {
			t.Errorf(test.NotSupported, suffix)
		}
	}

	if Supported("foo.wrong") {
		t.Errorf(test.Supported, ".wrong")
	}
}

func TestPathset(t *testing.T) {
	path, err := ioutil.TempDir("", "pathset-test")
	if err != nil {
		t.Fatalf("Error creating dir: %s", err)
	}
	defer os.RemoveAll(path)

	cases := []string{
		"foo.service",
		"foo.mount",
		"foo.socket",
		"foo.target",
		"foo.wrong",
	}

	correct := 0
	for _, name := range cases {
		if err = ioutil.WriteFile(filepath.Join(path, name), []byte{}, 0666); err != nil {
			t.Fatalf(test.ErrorIn, "ioutil.WriteFile", err)
		}

		if Supported(name) {
			correct++
		}
	}

	if paths, err := pathset(path); err != nil {
		t.Fatalf(test.ErrorIn, "pathset", err)
	} else if len(paths) != correct {
		t.Errorf(test.MismatchInVal, "len(paths)", len(paths), correct)
	}
}
