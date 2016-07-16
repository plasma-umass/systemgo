package init_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/b1101/systemgo/init"
	"github.com/b1101/systemgo/unit"
	"github.com/stretchr/testify/assert"
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
			After=b.service
			[Service]
			ExecStart=/bin/sleep 60`),
		},
		{
			"b.service",
			[]byte(`[Unit]
			Description=b.service
			RemainAfterExit=yes
			[Service]
			Type=oneshot
			ExecStart=/bin/echo test`),
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
			ExecStart=/bin/sleep 60`),
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
	var err error
	if dirpath, err = ioutil.TempDir("", "init-test"); err != nil {
		log.Fatalf("ioutil.Tempdir: %s", err)
	}

	etc = filepath.Join(dirpath, "etc", "systemd", "system")
	lib = filepath.Join(dirpath, "lib", "systemd", "system")
	Conf.Paths = []string{etc, lib}

	if err = populate(dirpath); err != nil {
		os.RemoveAll(dirpath)
		log.Fatalf("populate: %s", err)
	}
}

// TODO: Proper testing
func TestBoot(t *testing.T) {
	defer os.RemoveAll(dirpath)
	Boot()
	time.Sleep(time.Second)
	for _, units := range cases {
		for _, c := range units {
			u, err := System.Get(c.name)

			if !assert.NoError(t, err, c.name) {
				continue
			}

			u.Wait()
			fmt.Println(u.Status())
			assert.Equal(t, unit.Active, u.Active(), fmt.Sprintf("%s - %s", c.name, u.Active()))
		}
	}
}
