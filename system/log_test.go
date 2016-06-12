package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

var lorem = []byte(`
	Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	unc tortor magna, vestibulum a volutpat fermentum, aliquam quis enim.
	Quisque eget sapien nulla. Nulla et sem nec ante consequat auctor nec ut mauris.
	Praesent in nulla bibendum, sodales odio eget, posuere elit.
	Suspendisse tristique ligula non rutrum convallis.
	Maecenas eget urna ac nunc imperdiet tincidunt eu ut leo.
	Sed lacinia, ipsum et tincidunt viverra, metus ipsum pellentesque mi, quis laoreet nisl sapien quis nunc.
	Fusce aliquet metus sit amet libero euismod sollicitudin et ut arcu.`)

func TestRead(t *testing.T) {
	var err error

	var buf bytes.Buffer
	l := NewLog(&buf)

	if _, err = l.contents.ReadFrom(bytes.NewReader(lorem)); err != nil {
		t.Error(err)
	}

	var b []byte
	if b, err = ioutil.ReadAll(l); err != nil {
		t.Error(err)
	}

	defer func() {
		if t.Failed() {
			fmt.Println(string(b))
		}
	}()
	loremLines := len(bytes.Split(lorem, []byte("\n")))
	bLines := len(bytes.Split(b, []byte("\n")))
	if loremLines != bLines {
		t.Fatalf("Content string count mismatch should be: %v, encountered: %v", loremLines, bLines)
	}
}
