package system

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, len(lorem), l.Len(), "l.Len()")

	l.buffer.Reset()

	// []byte to write into buffer
	var b []byte

	// Times lorem fits in buffer
	var n int = BUFFER_SIZE / len(lorem)
	for i := 0; i <= n; i++ {
		b = append(b, lorem...)
	}

	n, err := l.Write(b)
	assert.NoError(t, err, "l.Write(b)")
	assert.Equal(t, n, BUFFER_SIZE, "l.Write(b)")
	assert.Equal(t, l.Cap(), BUFFER_SIZE, "l.Cap()")

	var char byte
	if (BUFFER_SIZE-len(b))%len(b) == 0 {
		char = b[0]
	} else {
		r := bytes.NewReader(b)
		for {
			// Search for first \n and char to next one after that
			c, err := r.ReadByte()
			require.NoError(t, err, "r.ReadByte()")

			if c == '\n' {
				char, _ = r.ReadByte()
				break
			}
		}
	}

	assert.Equal(t, char, l.buffer.Bytes()[0])
}

func TestRead(t *testing.T) {
	l := NewLog()
	l.buffer = bytes.NewBuffer(lorem)

	b, err := ioutil.ReadAll(l)
	assert.NoError(t, err, "first ioutil.ReadAll(l)")

	bTest, err := ioutil.ReadAll(l)
	assert.NoError(t, err, "second ioutil.ReadAll(l)")

	assert.Equal(t, b, bTest, "ioutil.ReadAll(l) bytes read")
}
