# Go Config

Package config provides support for configuration of go applications. It supports multiple configuration sources. 
Configuration variables can be defined only partially in a single source and multiple times in various sources.
Value for the same variable will be applied from the source with the highest priority - added as last to the slice 
of providers.

The following configuration sources are currently provided out of the box:

* yaml
* env

It is possible to customize which internal source should be used for configuration. Additional custom sources can be
configured and used with or without internal configuration providers.

### YAML

Configuration file is parsed using [go-yaml/yaml](https://github.com/go-yaml/yaml/tree/v3) module.
"config.yaml" file must be in the same directory as the application executable.  
Struct tags supported by the go-yaml/yaml module can be used.

### ENV
Environment variables should be named as uppercase field names, each nested struct name should
be inserted with an underscore ("_") prefix and postfix.  
Prefix of environment variables can be manually configured when env provider is initialized.  
Default configuration overwrites yaml configuration with values from environment.

## Usage

### Default configuration

```go
package main

import "github.com/tpodg/go-config"

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

import "github.com/tpodg/go-config"

func main() {
	cfg := struct {
		StringField string
		IntField    int
	}{}

	c := config.New()
	c.WithProviders(&pDummy{}, &config.Env{Prefix: "PREF"})

	err := c.Parse(&cfg)
	if err != nil {
		// handle error
	}
}

type pDummy struct{}

func (p *pDummy) Provide(config interface{}) error {
	// implementation
	return nil
}
```
