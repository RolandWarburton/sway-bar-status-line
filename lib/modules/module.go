package modules

type ModuleBehaviour interface {
	Run() error
	Init() error
}

type Module struct {
	ModuleBehaviour
	Enabled bool
}
