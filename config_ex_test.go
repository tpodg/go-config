package config_test

import "github.com/tpodg/go-config"

type pDummy struct{}

func (p *pDummy) Provide(config interface{}) error {
	return nil
}

func ExampleParse() {
	cfg := struct {
		param string
	}{}

	err := config.Parse(&cfg)
	if err != nil {
		// ...
	}
}

func ExampleC_WithProviders() {
	cfg := struct {
		param string
	}{}

	c := config.New()
	c.WithProviders(&pDummy{}, &config.Env{})

	err := c.Parse(&cfg)
	if err != nil {
		// ...
	}
}
