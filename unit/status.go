package unit

type Status struct {
	Load struct {
		Path   string `json:"Path"`
		Loaded Load   `json:"Loaded"`
		State  Enable `json:"Enabled"`
		Vendor struct {
			State int `json:"State"` // TODO: introduce a separate type
		} `json:"Vendor"`
	} `json:"Load"`

	Activation struct {
		State Activation `json:"State"`
		Sub   string     `json:"Sub"`
	} `json:"Activation"`

	Log []string `json:"Log,omitempty"`
}
