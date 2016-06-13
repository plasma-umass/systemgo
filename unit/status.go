package unit

import "fmt"

type Status struct {
	Load       LoadStatus       `json:"Load"`
	Activation ActivationStatus `json:"Activation"`

	Log []byte `json:"Log,omitempty"`
}
type ActivationStatus struct {
	State Activation `json:"State"`
	Sub   string     `json:"Sub"`
}
type LoadStatus struct {
	Path   string `json:"Path"`
	Loaded Load   `json:"Loaded"`
	State  Enable `json:"Enabled"`
	Vendor Enable `json:"Vendor"`
}

func (s Status) String() (out string) {
	defer func() {
		if len(s.Log) > 0 {
			out += fmt.Sprintf("\nLog:\n%s\n", s.Log)
		}
	}()
	return fmt.Sprintf(
		`Loaded: %s (%s; %s; vendor preset: %s)
Active: %s (%s)`,
		s.Load.Loaded, s.Load.Path, s.Load.State, s.Load.Vendor,
		s.Activation.State, s.Activation.Sub)
}
