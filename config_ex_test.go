package config_test

import "github.com/tpodg/go-config"

type pDummy struct{}

func (p *pDummy) Priority() int {
	return 100
}

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
	c.WithProviders(&config.Env{}, &pDummy{})

	err := c.Parse(&cfg)
	if err != nil {
		// ...
	}
}
