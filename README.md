# Config

### NOTE

**_This package is under active development and is not yet fully functional_**.

Package config provides support for configuration of go applications. It supports multiple configuration sources based
od priority.  
Configuration variables in a single source can be defined only partially and multiple times in various sources.
Configuration for the same variable will be applied from the source with the highest priority (1).

The following configuration sources are provided out of the box:

* yaml
* env

It is possible to customize which internal source will be used for configuration. Additional custom sources can be
configured and used with or without internal configuration providers.

### YAML

Configuration file is parsed with [go-yaml/yaml](https://github.com/go-yaml/yaml/tree/v3) module.
"config.yaml" file must be in the same directory as the application executable.  
Struct tags provided by the parsing module are supported. Priority of the provider is set to 50.

### ENV

**Not functional yet**  
Priority of the provider is set to 30. This way configuration from the yaml file can be overwritten with environment
variables.

## Usage

### Default configuration

```go
package main

import "github.com/tpodg/go-kit/config"

func main() {
	cfg := struct {
		StringField  string
		IntField     int
		NestedStruct struct {
			NestedString string
		}
	}{
		StringField: "string value",
	}

	err := config.Parse(&cfg)
	if err != nil {
		// handle error
	}
}
```

### Custom configuration

```go
package main

import "github.com/tpodg/go-kit/config"

func main() {
	cfg := struct {
		StringField string
		IntField    int
	}{}

	c := config.New()
	c.WithProviders(&config.Env{}, &pDummy{})

	err := c.Parse(&cfg)
	if err != nil {
		// handle error
	}
}

type pDummy struct{}

func (p *pDummy) Priority() int {
	return 100
}

func (p *pDummy) Provide(config interface{}) error {
	// implementation
	return nil
}
```
