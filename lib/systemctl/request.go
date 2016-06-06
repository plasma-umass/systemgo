package systemctl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Cmd   string
	Units []string
}

func NewRequest(cmd string, units ...string) (r *Request) {
	return &Request{cmd, units}
}

func (r *Request) Send(addr string) (body []byte, err error) {
	var b []byte
	if b, err = json.Marshal(r); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = http.Post(addr, "application/json", bytes.NewReader(b)); err != nil {
		return
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
