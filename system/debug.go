package system

import (
	"log"
	"os"
)

var Debug bool
var bug *log.Logger

func init() {
	bug = log.New(os.Stdout, "", log.Llongfile)
}
