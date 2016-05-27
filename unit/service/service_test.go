package service

import (
	"log"
	"testing"
)

func TestUnit(t *testing.T) {
	s := &Unit{}
	log.Println(s.Sub())
}
