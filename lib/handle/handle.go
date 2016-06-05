package handle

import (
	"errors"
	"log"
)

func Serr(str ...interface{}) {
	if len(str) > 0 {
		err := str[0].(string)
		str = str[0:]
		for _, st := range str {
			err += " " + st.(string)
		}
		Err(errors.New(err))
	}
	Err(errors.New("empty string specified in call to handle.Serr"))
}

func Err(err error) {
	log.Println(err.Error()) // TODO: Proper error handling
}
