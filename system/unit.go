package system

import (
	"bytes"
	"errors"
	"io"
	"log"
	"sync"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	Supervisable

	log UnitLog

	name string

	stats struct {
		path   string
		loaded unit.Load
	}

	listeners listeners
	rdy       chan interface{}
}

type listeners struct {
	ch []chan interface{}
	sync.Mutex
}

type UnitLog struct { // WIP
	*log.Logger
	out io.Writer
}

func NewUnit() (u *Unit) {
	var b bytes.Buffer
	u = &Unit{}
	u.log.Logger = log.New(&b, "", log.LstdFlags)
	u.log.out = &b
	u.rdy = make(chan interface{})
	go u.readyNotifier()
	return
}

func (u *Unit) readyNotifier() {
	for {
		<-u.rdy
		for _, c := range u.listeners.ch {
			c <- struct{}{}
			close(c)
		}
		u.listeners.ch = []chan interface{}{}
	}
}

func (u *Unit) ready() {
	u.rdy <- struct{}{}
}

func (u *Unit) waitFor() <-chan interface{} {
	u.listeners.Lock()
	c := make(chan interface{})
	u.listeners.ch = append(u.listeners.ch, c)
	u.listeners.Unlock()
	return c
}

func (u *Unit) Log(v ...interface{}) {
	str := ""
	if len(v) > 0 {
		str += v[0].(string)
		v = v[1:]

		for _, w := range v {
			str += " " + w.(string)
		}
	}
	u.log.Logger.Println(str)
}
func (u *Unit) SetOutput(w io.Writer) {
	u.log.Logger.SetOutput(w)
}

func (u Unit) Read(b []byte) (int, error) {
	if reader, ok := u.log.out.(io.Reader); ok {
		return reader.Read(b)
	}
	return 0, errors.New("unreadable")
}

func (u Unit) Name() string {
	return u.name
}
func (u Unit) Description() string {
	if u.Supervisable != nil {
		return u.Supervisable.Description()
	} else {
		return ""
	}
}

func (u Unit) Path() string {
	return u.stats.path
}
func (u Unit) Loaded() unit.Load {
	return u.stats.loaded
}

//func (u Unit) Enabled() unit.Enable {
//return u.stats.enabled
//}
func (u Unit) Active() unit.Activation {
	if u.Supervisable != nil {
		return u.Supervisable.Active()
	} else {
		return unit.Inactive
	}
}
