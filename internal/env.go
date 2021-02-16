package internal

type Env struct{}

func (e *Env) Priority() int {
	return 30
}

func (e *Env) Provide(config interface{}) error {
	// TODO
	return nil
}
