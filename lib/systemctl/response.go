package systemctl

// System response
import (
	"encoding/json"
	"errors"
	"io"

	"github.com/b1101/systemgo/system"
)

// Splits the system response into yield and error(if any)
func Parse(response io.Reader) (yield json.RawMessage, err error) {
	var resp struct {
		system.Response
		Yield json.RawMessage
	}
	if err = json.NewDecoder(response).Decode(&resp); err != nil {
		return
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return resp.Yield, err
}

// TODO: make a decoder
