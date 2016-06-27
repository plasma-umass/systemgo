package system

import (
	"github.com/b1101/systemgo/lib/errors"
	"github.com/b1101/systemgo/unit"
)

type Response struct {
	Yield interface{}
	Error string
}

type Handler func(...string) []Response

func (sys System) Handlers() map[string]Handler {
	return map[string]Handler{

		"status": func(names ...string) (resp []Response) {
			var err error

			if len(names) == 0 {
				resp = make([]Response, 1)
				if resp[0].Yield, err = sys.Status(); err != nil {
					resp[0].Error = err.Error()
				}
			} else {
				resp = make([]Response, len(names))
				for i, name := range names {
					if resp[i], err = sys.StatusOf(name); err != nil {
						resp[i].Error = err.Error()
					}
				}
			}
			return
		},
		"start": func(names ...string) (resp []Response) {
			resp = make([]Response, len(names))
			// WIP
			for i, name := range names {
				if err := sys.enqueue(name); err != nil {
					resp[i].Error = err.Error()
				}
			}
		},
		"stop": func(names ...string) (resp []Response) {
			return nil, sys.Stop(names)
		},
		"restart": func(names ...string) (resp []Response) {
			return nil, sys.Restart(names)
		},
		"reload": func(names ...string) (resp []Response) {
			return nil, sys.Reload(names)
		},
		"enable": func(names ...string) (resp []Response) {
			return nil, sys.Enable(names)
		},
		"disable": func(names ...string) (resp []Response) {
			return nil, sys.Disable(names)
		},
		"": func(names ...string) (resp []Response) {
			return nil, errors.WIP
		},
	}
}

func (sys *System) Start(names ...string) (err error) {
	if err = sys.NewJob(start, names...); err != nil {
		return
	}

	return job.Start()
}

func (sys *System) Stop(names ...string) (err error) {
	if err = sys.NewJob(stop, names...); err != nil {
		return
	}

	return job.Start()
}

func (sys *System) Restart(names ...string) (err error) {
	// rewrite as a loop to preserve errors
	if err = sys.Stop(names...); err != nil {
		return
	}
	return sys.Start(names...)
}

func (sys *System) Reload(name string) (err error) {
	if units, err := sys.units(names); err == nil {
		err = sys.reload(units)
	}
	return
	//var u *Unit
	//if u, err = sys.unit(name); err == nil {
	//if reloader, ok := u.Supervisable.(Reloader); ok {
	//reloader.Reload()
	//} else {
	//err = errors.NoReload
	//}
	//}
	//return
}
func (sys *System) Enable(name string) (err error) {
	if units, err := sys.units(names); err == nil {
		err = sys.enable(units)
	}
	return
	//var u *Unit
	//if u, err = sys.unit(name); err == nil {
	////sys.Enabled[u] = true
	//u.Enable()
	//}
	//return
	//}
}

func (sys *System) Disable(name string) (err error) {
	if units, err := sys.units(names); err == nil {
		err = sys.disable(units)
	}
	return
}

func (sys *System) StatusOf(name string) (statuses []unit.Status, err error) {
	statuses = make([]unit.Status, len(names))
	for i, name := range names {
		var u *Unit
		if statuses[i], err = sys.unit(name); err != nil {
			return
		}
	}
}

func (sys *System) Reload() (err error) {
	//
}
