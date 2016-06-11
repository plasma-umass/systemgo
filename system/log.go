package system

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type Log struct {
	*log.Logger
	out io.Writer
	contents
}

func NewLog(w io.Writer) *Log {
	return &Log{
		Logger: log.New(w, "", log.LstdFlags),
		out:    w,
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

	if n, err = reader.Read(b); err != nil && err != io.EOF {
		return
	}

	if err = func() error {
		nWr, errWr := l.contents.Write(b)
		if errWr == nil && nWr != n {
			return fmt.Errorf("Byte count mismatch: %v read, %v written", n, nWr)
		}
		return errWr
	}(); err != nil {
		err = fmt.Errorf("Error writing contents: %v", err)
	}

	return
}

type contents struct {
	Contents  []string
	byteCount int
}

func (c contents) Len() int {
	return c.byteCount
}

func (c *contents) Write(b []byte) (n int, err error) {
	data := strings.Split(string(b), "\n")

	defer func() {
		c.byteCount += n
	}()
	countBytes := func(arr []string) int {
		var n int
		for _, str := range arr {
			n += len(str)
		}
		return n - 1
	}

	switch {
	case len(data) == 0:
		return 0, fmt.Errorf("Empty array received")
	case len(c.Contents)+len(data) <= cap(c.Contents):
		c.Contents = append(c.Contents, data...)
		return len(b), nil
	case len(data) >= cap(c.Contents):
		c.byteCount = 0
		c.Contents = data[len(data)-cap(c.Contents):]
		return countBytes(c.Contents), fmt.Errorf("Length of data exceeds capacity: %v > %v", len(data), cap(c.Contents))
	default:
		var offset int
		if offset = cap(c.Contents) - len(data); offset < 0 {
			offset = 0
		}

		c.byteCount = countBytes(c.Contents[offset:])

		c.Contents = append(c.Contents[offset:], data...)
		return len(b), nil
	}
}
