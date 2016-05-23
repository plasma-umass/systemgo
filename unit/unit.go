package unit

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

var supported = map[string]bool{
	"simple": true,
}

const MAXLENGTH = 2048

// Struct representing the unit
type Unit struct {
	*Definition
	*exec.Cmd
	Loaded bool
	Status status
	//Name      string
	//ShortName string
}

// Definition as found in the unit specification file
type Definition struct {
	Unit struct {
		Description, After, Wants string
	}
	Service struct {
		//Type                        ServiceType
		Type, ExecStart, WorkingDirectory string
	}
	Install struct {
		// WIP
	}
}

//type ServiceType string

//const (
//Simple  ServiceType = ""
//Oneshot             = "oneshot"
//Forking             = "forking"
//)

// Execution status of a unit
type status int

const (
	Inactive status = iota
	Active
	Exited
	Failed
)

// Looks for unit files in paths given and adds to 'Units' map
func ParseDir(paths ...string) (map[string]*Unit, error) {
	units := map[string]*Unit{}
	for _, path := range paths {
		var err error
		var file *os.File
		var files []os.FileInfo

		if file, err = os.Open(path); err != nil {
			return nil, err
		}
		if files, err = file.Readdir(0); err != nil {
			if err == io.EOF {
				log.Println(path + " is empty")
				err = nil
				break
			}
			return nil, err
		}

		for _, f := range files {
			var def *Definition

			fpath := path + "/" + f.Name()

			if f.IsDir() {
				//return ParseDir(fpath)
				//			return nil, errors.New("not implemented yet") //TODO
				log.Println("directory parsing not implemented yet")
				continue
			}
			if file, err = os.Open(fpath); err != nil {
				return nil, err
			}
			if def, err = ParseUnit(file); err != nil {
				//return nil, errors.New("error parsing " + fpath + ": " + err.Error())
				log.Println("error parsing " + fpath + ": " + err.Error())
				continue
			}

			units[f.Name()] = &Unit{Definition: def}
		}
	}
	return units, nil
}

// Attempts to parse a specification of a unit
func ParseUnit(specification io.Reader) (*Definition, error) {
	var err error
	var opts []*unit.UnitOption
	definition := &Definition{}
	def := reflect.ValueOf(definition).Elem()

	if opts, err = unit.Deserialize(specification); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if v := def.FieldByName(opt.Section); v.IsValid() {
			if v := v.FieldByName(opt.Name); v.IsValid() {
				v.SetString(opt.Value)
			} else {
				return nil, errors.New("field " + opt.Name + " does not exist")
			}
		} else {
			return nil, errors.New("section " + opt.Section + " does not exist")
		}
	}

	if definition.Service.ExecStart == "" {
		return nil, errors.New("'ExecStart' field not set")
	}

	switch definition.Service.Type {
	case "":
		definition.Service.Type = "simple"
	default:
		if !supported[definition.Service.Type] {
			return nil, errors.New("service type " + definition.Service.Type + " does not exist or is not supported yet")
		}
	}

	return definition, nil
}

// Starts execution of the unit's specified command
func (u *Unit) Start() error {
	cmd := strings.Split(u.Service.ExecStart, " ")
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	u.Loaded = true
	return u.Cmd.Start()
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() error {
	if u.Loaded != true {
		return errors.New("unit not loaded")
	}

	u.Loaded = false
	return u.Process.Kill()
}
