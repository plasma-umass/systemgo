package unit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
)

var (
	// Test cases
	tests = map[string][]struct {
		name     string
		correct  bool
		contents string
	}{
		"test1": {
			{
				"test1.service", true,
				`[Unit]
				Description=test service 1
				[Service]
				ExecStart=echo test 1	
				`,
			},
			{
				"override.service", true,
				`[Unit]
				Description=Not overriden
				[Service]
				ExecStart=echo test 2
				`,
			},
		},
		"test2": {
			{
				"override.service", true,
				`[Unit]
				Description=Overriden
				[Service]
				ExecStart=echo test 2	
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
	for dir, units := range tests {
		if err := os.Mkdir(dir, 0755); err != nil {
			return errors.New("Failed to create directory: " + err.Error())
		}
		for _, unit := range units {
			if u, err := os.Create(dir + "/" + unit.name); err != nil {
				return errors.New("Failed to create file: " + err.Error())
			} else {
				defer u.Close()
				if _, err := u.Write([]byte(unit.contents)); err != nil {
					return errors.New("Failed to write contents: " + err.Error())
				}
			}
		}
	}
	return nil
}

func RemoveUnits() error {
	for dir, _ := range tests {
		if err := os.RemoveAll(dir); err != nil {
			return errors.New("Failed to remove directory " + dir + ": " + err.Error())
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
	if err := CreateUnits(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := ParseDir(paths...); err != nil {
		log.Println(err.Error())
	}

	for _, dir := range tests {
		for _, unit := range dir {
			if _, ok := Units[unit.name]; ok != unit.correct {
				t.Error(unit.name, "should be", unit.correct)
			}
		}
	}
	u, _ := Units["override.service"]
	if u.Unit.Description != "Overriden" {
		t.Error("Unit file specification was not overriden")
	}
	if err := RemoveUnits(); err != nil {
		log.Fatalln(err.Error())
	}
}
