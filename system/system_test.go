package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

var cases = map[string][]struct {
	name     string
	contents []byte
}{
	"etc/systemd/system": {
		{
			"a.service",
			[]byte(`[Unit]
			Description=a.service
			[Service]
			ExecStart=/bin/sleep 2`),
		},
		{
			"b.service",
			[]byte(`[Unit]
			Description=b.service
			[Service]
			Type=oneshot
			ExecStart=/bin/sleep 2`),
		},
		{
			"c.target",
			[]byte(`[Unit]
			Description=c.target
			Requires=a.service b.service
			`),
		},
	},
	"lib/systemd/system": {
		{
			"d.service",
			[]byte(`[Unit]
			Description=d.service
			[Service]
			ExecStart=/bin/sleep 2`),
		},
	},
}

func populate(path string) (err error) {
	for dir, units := range cases {
		dirpath := filepath.Join(path, dir)
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return fmt.Errorf("os.MkdirAll(%s, 0755): %s", dirpath, err)
		}
		for _, unit := range units {
			fpath := filepath.Join(dirpath, unit.name)

			file, err := os.Create(fpath)
			if err != nil {
				return fmt.Errorf("os.Create(%s)", fpath, err)
			}
			defer file.Close()

			if _, err = file.Write(unit.contents); err != nil {
				return fmt.Errorf("%s - file.Write: %s", file.Name(), err)
			}
		}
	}
	return
}

func TestLoad(t *testing.T) {
	loaded := 0

	sys := New()

	testPtr := func(name string, ptr *Unit) {
		if u, _ := sys.Get(name); u != ptr {
			t.Errorf(test.MismatchInVal, "sys.Get("+name+")", u, ptr)
		}
	}

	path, err := ioutil.TempDir("", "system-test")
	if err != nil {
		log.Fatalf("ioutil.TempDir: %s", err)
	}
	defer os.RemoveAll(path)

	if err = os.Chdir(path); err != nil {
		log.Fatalf("os.Chdir(%s): %s", path, err)
	}

	if err = populate(path); err != nil {
		log.Fatalf("populate(%s): %s", path, err)
	}

	paths := make([]string, 0, len(cases))
	for p := range cases {
		paths = append(paths, filepath.Join(path, p))
	}
	sys.SetPaths(paths...)

	for _, units := range cases {
		for _, unit := range units {
			if ptr, err := sys.Get(unit.name); err != nil {
				t.Errorf(test.ErrorIn, "sys.Get("+unit.name+")", err)
			} else {
				loaded++
				defer testPtr(unit.name, ptr)
			}
		}
	}

	fpath := filepath.Join(os.TempDir(), "link.target")

	err = ioutil.WriteFile(fpath, []byte(`[Unit]
Description=linked unit`), 0666)
	if err != nil {
		log.Fatalf(test.ErrorIn, "ioutil.WriteFile", err)
	}

	if ptr, err := sys.Load(fpath); err != nil {
		t.Errorf(test.ErrorIn, "sys.Load("+fpath+")", err)
	} else {
		defer testPtr(fpath, ptr)
	}

	loaded++

	if len(sys.units) != loaded {
		t.Errorf(test.MismatchInVal, "len(sys.units)", len(sys.units), loaded)
	}
}
