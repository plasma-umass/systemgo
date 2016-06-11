package systemctl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

// System communication client using HTTP protocol
type HttpClient struct {
	addr string
}

func NewHttpClient(addr string) *HttpClient {
	return &HttpClient{
		addr,
	}
}

func (c HttpClient) Addr() string {
	return c.addr
}
func (c HttpClient) String() string {
	return c.Addr()
}

func (c *HttpClient) Get(cmd string, units ...string) (dec *json.Decoder, err error) {
	var resp *http.Response
	if resp, err = getHTTP(c.Addr(), cmd, units...); err != nil {
		return
	}
	defer resp.Body.Close()

	var yield json.RawMessage
	if yield, err = Parse(resp.Body); err != nil {
		return
	}

	return json.NewDecoder(bytes.NewReader(yield)), nil
}

func (c *HttpClient) Do(cmd string, units ...string) (err error) {
	var resp *http.Response
	if resp, err = getHTTP(c.Addr(), cmd, units...); err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = Parse(resp.Body)

	return
}

func getHTTP(addr, cmd string, units ...string) (*http.Response, error) {
	if len(units) == 0 {
		return http.Get(addr + "/" + cmd + "/")
	} else {
		v := url.Values{}
		for _, unit := range units {
			v.Add("unit", unit)
		}

		return http.Get(addr + "/" + cmd + "?" + v.Encode())
	}
}
