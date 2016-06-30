package init

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/b1101/systemgo/system"
	"github.com/b1101/systemgo/test"
)

var units = map[string][]struct {
	name     string
	contents []byte
}{
	"etc/systemd/system": {
		{
			"a.service",
			[]byte(`[Unit]
			Description=a.service
			After=b.service
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
			"default.target",
			[]byte(`[Unit]
			Description=c.target
			Requires=a.service b.service
			Wants=d.service
			After=a.service
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
	for dir, units := range units {
		dirpath := filepath.Join(path, dir)
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return fmt.Errorf("os.MkdirAll(%s, 0755): %s", dirpath, err)
		}
		for _, unit := range units {
			fpath := filepath.Join(dirpath, unit.name)

			file, err := os.Create(fpath)
			if err != nil {
				return fmt.Errorf("os.Create(%s):%s", fpath, err)
			}
			defer file.Close()

			if _, err = file.Write(unit.contents); err != nil {
				return fmt.Errorf("%s - file.Write: %s", file.Name(), err)
			}
		}
	}
	return
}

var etc, lib, dirpath string

func init() {
	system.Debug = true

	var err error
	if dirpath, err = ioutil.TempDir("", "init-test"); err != nil {
		log.Fatalf(test.ErrorIn, "ioutil.Tempdir", err)
	}

	etc = filepath.Join(dirpath, "etc", "systemd", "system")
	lib = filepath.Join(dirpath, "lib", "systemd", "system")
	conf.Paths = []string{etc, lib}

	if err = populate(dirpath); err != nil {
		os.RemoveAll(dirpath)
		log.Fatalf(test.ErrorIn, "populate", err)
	}
}

func TestBoot(t *testing.T) {
	defer os.RemoveAll(dirpath)
	Boot()
	time.Sleep(5 * time.Second)
}
