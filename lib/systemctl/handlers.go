package systemctl

import "github.com/b1101/systemgo/system"

type Handler func(...string) []Response

func Handlers(sys system.Daemon) map[string]Handler {
	return map[string]Handler{
		"status": func(names ...string) (resp []Response) {
			if len(names) == 0 {
				return handle(func() (interface{}, error) {
					return sys.Status()
				})
			} else {
				return handleEach(names, func(name string) (interface{}, error) {
					return sys.StatusOf(name)
				})
			}
			return
		},
		"start": func(names ...string) (resp []Response) {
			return handleErr(func() error {
				return sys.Start(names...)
			})
		},
		"stop": func(names ...string) (resp []Response) {
			return handleEachErr(names, func(name string) error {
				return sys.Stop(name)
			})
		},
		"restart": func(names ...string) (resp []Response) {
			return handleEachErr(names, func(name string) error {
				return sys.Restart(name)
			})
		},
		"reload": func(names ...string) (resp []Response) {
			return handleEachErr(names, func(string) error {
				//return sys.Reload(name)
				return system.ErrNotImplemented // TODO
			})
		},
		"enable": func(names ...string) (resp []Response) {
			return handleEachErr(names, func(name string) error {
				//return sys.Enable(name)
				return system.ErrNotImplemented // TODO
			})
		},
		"disable": func(names ...string) (resp []Response) {
			return handleEachErr(names, func(name string) error {
				//return sys.Disable(name)
				return system.ErrNotImplemented // TODO
			})
		},
		"": func(names ...string) (resp []Response) {
			return handle(func() (interface{}, error) {
				//return sys.ListUnits()
				return nil, system.ErrNotImplemented // TODO
			})
		},
	}
}
func handle(fn func() (interface{}, error)) (resp []Response) {
	resp = make([]Response, 1)

	var err error
	if resp[0].Yield, err = fn(); err != nil {
		resp[0].Error = err.Error()
	}
	return
}

func handleEach(names []string, fn func(string) (interface{}, error)) (resp []Response) {
	resp = make([]Response, len(names))

	for i, name := range names {
		var err error
		if resp[i].Yield, err = fn(name); err != nil {
			resp[i].Error = err.Error()
		}
	}
	return
}

func handleErr(fn func() error) (resp []Response) {
	resp = make([]Response, 1)

	if err := fn(); err != nil {
		resp[0].Error = err.Error()
	}
	return
}

func handleEachErr(names []string, fn func(string) error) (resp []Response) {
	resp = make([]Response, len(names))

	for i, name := range names {
		if err := fn(name); err != nil {
			resp[i].Error = err.Error()
		}
	}
	return
}
