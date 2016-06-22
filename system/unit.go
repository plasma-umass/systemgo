package system

import (
	"sync"

	"github.com/b1101/systemgo/unit"
)

type Unit struct {
	Supervisable

	path   string
	loaded unit.Load

	Requires, Wants, After, Before []*Unit

	listeners listeners
	rdy       chan interface{}

	Log *Log
}

type listeners struct {
	ch []chan interface{}
	sync.Mutex
}

func NewUnit() (u *Unit) {
	defer func() {
		go u.readyNotifier()
	}()
	return &Unit{
		Log: NewLog(),
		rdy: make(chan interface{}),
	}
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

func (u Unit) Path() string {
	return u.path
}
func (u Unit) Loaded() unit.Load {
	return u.loaded
}
func (u Unit) Description() string {
	if u.Supervisable == nil {
		return ""
	}

	return u.Supervisable.Description()
}
func (u Unit) Active() unit.Activation {
	if u.Supervisable == nil {
		return unit.Inactive
	}

	if subber, ok := u.Supervisable.(unit.Subber); ok {
		return subber.Active()
	}

	for _, dep := range u.Requires { // TODO: find out what systemd does
		if dep.Active() != unit.Active {
			return unit.Inactive
		}
	}

	return unit.Active
}
func (u Unit) Sub() string {
	if u.Supervisable == nil {
		return "dead"
	}

	if subber, ok := u.Supervisable.(unit.Subber); ok {
		return subber.Sub()
	}

	return u.Active().String()
}
