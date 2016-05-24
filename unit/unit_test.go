package unit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	// Test cases
	tests = map[string][]struct {
		name     string
		correct  bool
		contents string
	}{
		"lib/systemd/system": {
			{
				"test1.service", true,
				`[Unit]
			Description=test service 1
			Requires=override.service
			After=override.service test4.service
			[Service]
			ExecStart=echo test 1	
				`,
			},
			{
				"override.service", true,
				`[Unit]
			Description=Not overriden
			Requires=test3.service
			[Service]
			ExecStart=echo test 2
				`,
			},
		},
		"etc/systemd/system": {
			{
				"override.service", true,
				`[Unit]
			Description=Overriden
			Wants=test4.service
			Requires=test1.service
			[Service]
			ExecStart=echo test 2	
			Restart=yes
				`,
			},
			{
				"test3.service", false,
				`[Unit]
			Description=test service 3
			FooBar=foo
			[Service]
			ExecStart=echo test 3
				`,
			},
			{
				"test4.service", false,
				`[Unit]
			Description=test service 3
			[Service]
			Type="foobar"
			ExecStart=echo test 3
				`,
			},
		},
	}
	paths = []string{}
)

func init() {
	for path := range tests {
		paths = append(paths, path)
	}
}

func CreateUnits() error {
	for path, units := range tests {
		if err := os.MkdirAll(path, 0755); err != nil {
			return errors.New("Failed to create" + path + ": " + err.Error())
		}
		for _, unit := range units {
			if file, err := os.Create(path + "/" + unit.name); err != nil {
				return errors.New("Failed to create" + path + "/" + unit.name + ": " + err.Error())
			} else {
				defer file.Close()
				if _, err := file.Write([]byte(unit.contents)); err != nil {
					return errors.New("Failed to write contents to  " + file.Name() + ": " + err.Error())
				}
			}
		}
	}
	return nil
}

func RemoveUnits() error {
	for path, _ := range tests {
		if err := os.RemoveAll(strings.Split(path, "/")[0]); err != nil {
			return errors.New("Failed to remove" + path + ": " + err.Error())
		}
	}
	return nil
}

func ExampleStatus() {
	var s status = Active
	fmt.Println(s)
	// Output:
	// Active
}

func TestParse(t *testing.T) {
	var err error
	var units map[string]*Unit
	if err = CreateUnits(); err != nil {
		log.Println("error")
		if er := RemoveUnits(); err != nil {
			log.Println(er.Error())
		}
		log.Fatalln(err.Error())
	}
	if units, err = ParseDir(paths...); err != nil {
		log.Println(err.Error())
	}

	for _, dir := range tests {
		for _, unit := range dir {
			if _, ok := units[unit.name]; ok != unit.correct {
				t.Error(unit.name, "should be", unit.correct)
			}
		}
	}
	u, ok := units["override.service"]
	if !ok {
		t.Error("override.service is not there")
		return
	}
	if u.Unit.Description != "Overriden" {
		t.Error("Unit file specification was not overriden")
	}
	if err := RemoveUnits(); err != nil {
		log.Fatalln(err.Error())
	}
}

func TestStart(t *testing.T) {
	var err error
	var units map[string]*Unit
	if err = CreateUnits(); err != nil {
		log.Fatalln(err.Error())
	}
	if units, err = ParseDir(paths...); err != nil {
		log.Println(err.Error())
	}

	Units = units
	Loaded = map[*Unit]bool{}

	for _, u := range units {
		if err = u.Start(); err != nil {
			log.Println(err.Error())
		}
	}
	if err := RemoveUnits(); err != nil {
		log.Fatalln(err.Error())
	}
}
