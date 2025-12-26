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

Configuration file is parsed using [goccy/go-yaml](https://github.com/goccy/go-yaml) module.
"config.yaml" file must be in the same directory as the application executable by default.  
Custom path for the configuration file can be set using the `Path` field of the `Yaml` provider. If a relative path is provided, it will be resolved relative to the application's executable directory.
Struct tags supported by the goccy/go-yaml module can be used.

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
	c.WithProviders(&config.Yaml{Path: "/path/to/config.yaml"}, &config.Env{Prefix: "PREF"})

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
