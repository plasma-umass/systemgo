package unit

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/unit"
)

var supported = map[string]bool{
	"simple": true,
}

var (
	Units  map[string]*Unit
	Loaded map[*Unit]bool
)

const MAXLENGTH = 2048

// Struct representing the unit
type Unit struct {
	*Definition
	*exec.Cmd
	Status status
	State  state
	//Name      string
	//ShortName string
}

// Definition as found in the unit specification file
type Definition struct {
	Unit struct {
		Description                              string
		After, Wants, Requires, Conflicts, Before []string
	}
	Service struct {
		//Type                        ServiceType
		Type, ExecStart, ExecReload, WorkingDirectory string
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

type state int

const (
	Static state = iota
	Indirect
	Disabled
	Enabled
)

// Looks for unit files in paths given and returns a map of units
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
				log.Println("recursive directory parsing not implemented yet")
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
				switch v.Kind() {
				case reflect.SliceOf(reflect.TypeOf(reflect.String)).Kind():
					values := strings.Split(opt.Value, " ")
					v.Set(reflect.ValueOf(values))
				case reflect.SliceOf(reflect.TypeOf(reflect.Int)).Kind():
					values := strings.Split(opt.Value, " ")
					ints := make([]int, len(values), len(values))
					for i, val := range values {
						if ints[i], err = strconv.Atoi(val); err != nil {
							return nil, errors.New("Error parsing " + opt.Name + ": " + err.Error())
						}
					}
					v.Set(reflect.ValueOf(ints))
				default:
					v.SetString(opt.Value)
				}
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

	if definition.Service.Type == "" {
		definition.Service.Type = "simple"
	}

	if !supported[definition.Service.Type] {
		return nil, errors.New("service type " + definition.Service.Type + " does not exist or is not supported yet")
	}

	return definition, nil
}

// Starts execution of the unit's specified command
func (u *Unit) Start() (err error) {
	cmd := strings.Split(u.Service.ExecStart, " ")
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	for _, name := range u.Definition.Unit.Conflicts {
		if dep, ok := Units[name]; ok {
			if Loaded[dep] {
				return errors.New("conflicts with " + name)
			}
		}
	}
	for _, name := range u.Definition.Unit.Requires {
		if dep, ok := Units[name]; !ok {
			return errors.New(name + " not found")
		} else {
			if !Loaded[dep] {
				if err = dep.Start(); err != nil {
					return
				}
			}
			if dep.GetStatus() != Active {
				return errors.New(name + " failed to launch")
			}
		}
	}
	for _, name := range u.Definition.Unit.Wants {
		if dep, ok := Units[name]; !ok {
			return errors.New(name + " not found")
		} else {
			if !Loaded[dep] {
				if err = dep.Start(); err != nil {
					return
				}
			}
		}
	}
	for _, name := range u.Definition.Unit.After {
		if dep, ok := Units[name]; !ok {
			return errors.New(name + " not found")
		} else {
			for !Loaded[dep] {
				time.Sleep(time.Second)
			}
		}
	}

	switch u.Service.Type {
	case "simple":
		err = u.Cmd.Start()
	case "oneshot":
		err = u.Cmd.Run()
	case "forking", "dbus", "notify", "idle":
		return errors.New(u.Service.Type + " not implemented yet") // TODO
	default:
		return errors.New(u.Service.Type + " does not exist")
	}
	Loaded[u] = true

	return
}

// Stops execution of the unit's specified command
func (u *Unit) Stop() (err error) {
	if !Loaded[u] {
		return errors.New("unit not loaded")
	}

	err = u.Process.Kill()
	//u.Loaded = false
	delete(Loaded, u)
	return
}

// Stop and restart unit execution
func (u *Unit) Restart() (err error) {
	if err = u.Stop(); err != nil {
		return
	}
	delete(Loaded, u)

	var cmd []string
	if u.Service.ExecReload == "" {
		cmd = strings.Split(u.Service.ExecStart, " ")
	} else {
		cmd = strings.Split(u.Service.ExecReload, " ")
	}
	u.Cmd = exec.Command(cmd[0], strings.Join(cmd[1:], " "))

	err = u.Cmd.Start()
	//u.Loaded = true
	Loaded[u] = true
	return
}

// Reload unit definition
func (u *Unit) Reload() error {
	//u.Definition, _ = ParseUnit(u)
	return errors.New("not implemented yet") // TODO
}

func (u *Unit) GetStatus() status {
	// TODO: Fixme, check proper state
	switch {
	case u.Cmd.ProcessState == nil:
		switch u.Status {
		case Failed:
			return Failed
		default:
			return Inactive
		}
	case u.Cmd.ProcessState.Success():
		switch u.Service.Type {
		case "oneshot":
			return Active
		case "simple":
			return Exited
		default:
			return Failed
		}
	case u.Cmd.ProcessState.Exited():
		return Failed
	default:
		return Active
	}
}
