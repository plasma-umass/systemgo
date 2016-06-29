package unit

//go:generate mockgen -source=interfaces.go -package=unit -destination=mock_interfaces_test.go
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
