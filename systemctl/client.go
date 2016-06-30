package systemctl

import "encoding/json"

type Client interface {
	// Performs the system request,
	// decodes error received, returns a decoder containing the yield
	Get(string, ...string) (*json.Decoder, error)

	// Performs the system request,
	// decodes error received, discards everything else
	Do(string, ...string) error
}
