package unit

import "io"

type Definer interface {
	Define(io.Reader) error
}

// Subber is implemented by any value that has Sub and Active methods
type Subber interface {
	Active() Activation
	Sub() string
}

// StartStopper is implemented by any value that has Start and Stop methods
type StartStopper interface {
	Starter
	Stopper
}

// Starter is implemented by any value that has a Start method
type Starter interface {
	Start() error
}

// Stopper is implemented by any value that has a Stop method
type Stopper interface {
	Stop() error
}

// Reloader is implemented by any value capable of reloading itself(or its definition)
type Reloader interface {
	Reload() error
}
