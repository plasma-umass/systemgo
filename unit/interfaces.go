package unit

type Subber interface {
	Active() Activation
	Sub() string
}

type StartStopper interface {
	Starter
	Stopper
}
type Starter interface {
	Start() error
}
type Stopper interface {
	Stop() error
}
type Reloader interface {
	Reload() error
}
