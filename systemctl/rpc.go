package systemctl

import (
	"encoding/gob"
	"fmt"

	"github.com/rvolosatovs/systemgo/unit"
)

type Response struct {
	Yield interface{}
}

func init() {
	gob.Register(map[string]unit.Status{})
}

func newResponse() (resp *Response) {
	return &Response{
		Yield: map[string]fmt.Stringer{},
	}
}

func NewServer(sys Daemon) (sv *Server) {
	return &Server{sys}
}

type Server struct {
	sys Daemon
}

func (sv *Server) Start(names []string, resp *Response) (err error) {
	return sv.sys.Start(names...)
}

func (sv *Server) Stop(names []string, resp *Response) (err error) {
	return sv.sys.Stop(names...)
}

func (sv *Server) Restart(names []string, resp *Response) (err error) {
	return sv.sys.Restart(names...)
}

func (sv *Server) Isolate(names []string, resp *Response) (err error) {
	return sv.sys.Isolate(names...)
}

func (sv *Server) Reload(names []string, resp *Response) (err error) {
	return sv.sys.Reload(names...)
}

func (sv *Server) Enable(names []string, resp *Response) (err error) {
	return sv.sys.Enable(names...)
}

func (sv *Server) Disable(names []string, resp *Response) (err error) {
	return sv.sys.Disable(names...)
}

func (sv *Server) Status(names []string, resp *Response) (err error) {
	*resp = *newResponse()

	statuses := map[string]unit.Status{}

	for _, name := range names {
		var st unit.Status
		if st, err = sv.sys.StatusOf(name); err != nil {
			continue
		}

		statuses[name] = st
	}

	resp.Yield = statuses
	return err
}

func (sv *Server) StatusAll(names []string, resp *Response) (err error) {
	units := sv.sys.Units()

	names = make([]string, 0, len(units))
	for _, u := range units {
		names = append(names, u.Name())
	}
	return sv.Status(names, resp)
}
