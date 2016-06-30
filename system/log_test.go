package system

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/b1101/systemgo/test"
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

func TestWrite(t *testing.T) {
	l := NewLog()
	l.Write(lorem)

	if len(lorem) != l.Len() {
		t.Errorf(test.MismatchInVal, "l.Len()", l.Len(), len(lorem))
	}

	l.buffer.Reset()

	// []byte to write into buffer
	var b []byte

	// Times lorem fits in buffer
	var n int = BUFFER_SIZE / len(lorem)
	for i := 0; i <= n; i++ {
		b = append(b, lorem...)
	}

	if n, err := l.Write(b); err != nil {
		t.Errorf(test.ErrorIn, "l.Write(b)", err)
	} else if n != BUFFER_SIZE {
		t.Errorf(test.MismatchInVal, "l.Write(b)", n, BUFFER_SIZE)
	}

	if l.Cap() != BUFFER_SIZE {
		t.Errorf(test.MismatchInVal, "l.Cap()", l.Cap(), BUFFER_SIZE)
	}

	var char byte
	if (BUFFER_SIZE-len(b))%len(b) == 0 {
		char = b[0]
	} else {
		r := bytes.NewReader(b)
		for {
			// Search for first \n and char to next one after that
			if c, err := r.ReadByte(); err != nil {
				t.Fatalf(test.ErrorIn, "r.ReadByte()", err)
			} else if c == '\n' {
				char, _ = r.ReadByte()
				break
			}
		}
	}

	if l.buffer.Bytes()[0] != char {
		t.Errorf(test.MismatchIn, "l.buffer.Bytes()[0]", string(l.buffer.Bytes()[0]), string(char))
	}
}

func TestRead(t *testing.T) {
	l := NewLog()
	l.buffer = bytes.NewBuffer(lorem)

	b, err := ioutil.ReadAll(l)
	if err != nil {
		t.Error(err)
	}

	bTest, err := ioutil.ReadAll(l)
	if err != nil {
		t.Errorf(test.Error, err)
	} else if string(bTest) != string(b) {
		t.Errorf(test.MismatchInVal, "ioutil.ReadAll(l)", bTest, b)
	}
}
