package system

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type Log struct {
	*log.Logger
	out      io.Writer
	contents contents
}

func NewLog(w io.Writer) *Log {
	return &Log{
		Logger: log.New(w, "", log.LstdFlags),
		out:    w,
		contents: contents{
			Buffer: bytes.NewBuffer(make([]byte, 0, 10000)),
		},
	}
}

func (l *Log) SetOutput(w io.Writer) {
	l.out = w
	l.Logger.SetOutput(w)
}

func (l *Log) Read(b []byte) (n int, err error) {
	var reader io.Reader
	var ok bool
	if reader, ok = l.out.(io.Reader); !ok {
		return 0, fmt.Errorf("Cannot read from %T", l.out)
	}

	if _, err = l.contents.ReadFrom(reader); err != nil {
		return
	}

	return l.contents.Read(b)
}

type contents struct {
	*bytes.Buffer
	io.Reader
}

func (c *contents) Read(b []byte) (n int, err error) {
	if c.Reader == nil {
		c.Reader = bytes.NewReader(c.Buffer.Bytes())
	}
	defer func() {
		if err == io.EOF {
			c.Reader = nil
		}
	}()
	return c.Reader.Read(b)
}

func (c *contents) ReadFrom(r io.Reader) (n int64, err error) {
	var b []byte
	if b, err = ioutil.ReadAll(r); err != nil {
		return
	}
	_, err = c.Write(b)
	return int64(len(b)), err
}

func (c *contents) Write(b []byte) (n int, err error) {
	if c.Len()+len(b) <= c.Cap() {
		return c.Buffer.Write(b)
	}

	defer func() {
		if err == nil {
			_, err = c.Buffer.ReadString('\n')
		}
	}()

	if len(b) >= c.Cap() {
		c.Buffer.Reset()
		return c.Buffer.Write(b[len(b)-c.Cap():])
	}

	if _, err = c.Buffer.Read(make([]byte, len(b)+c.Len()-c.Cap())); err != nil {
		return 0, err
	}

	return c.Buffer.Write(b)
}
