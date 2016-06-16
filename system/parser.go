package system

import "log"

type parser struct {
	*log.Logger
	paths  []string
	loaded map[string]*Supervisable
}


