package systemctl

import "fmt"

type QueryResponse struct {
	Yield interface{}
	Errs  ExecResponse
}

type ExecResponse map[string]error

func NewServer(sys Daemon) (sv *Server) {
	return &Server{sys}
}

type Server struct {
	sys Daemon
}

func (sv *Server) Start(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Start(names...)
}

func (sv *Server) Stop(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Stop(names...)
}

func (sv *Server) Restart(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Restart(names...)
}

func (sv *Server) Isolate(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Isolate(names...)
}

func (sv *Server) Reload(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Reload(names...)
}

func (sv *Server) Enable(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Enable(names...)
}

func (sv *Server) Disable(names []string, errs *ExecResponse) (err error) {
	return sv.sys.Disable(names...)
}

func (sv *Server) Status(names []string, resp *QueryResponse) (err error) {
	//TODO
	//statuses := map[string]system.Status
	statuses := map[string]fmt.Stringer{}

	for _, name := range names {
		statuses[name], resp.Errs[name] = sv.sys.StatusOf(name)
	}

	resp.Yield = statuses
	return nil
}
