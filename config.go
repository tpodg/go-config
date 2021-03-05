// Package config provides support for application configuration. Configuration values can be
// provided from multiple sources / providers. If the same value is configured in multiple
// sources, value from the source with the highest priority (added last to the slice of providers) will be applied.
package config

import (
	"errors"
	"reflect"
)

// Provider interface offers support for multiple configuration sources.
type Provider interface {

	// Provide is an interface for specific configuration source implementation. It reads values from
	// the configuration source and maps them to the configuration struct.
	Provide(config interface{}) error
}

// C is a wrapper struct holding a slice of configuration sources, which must implement Provider interface.
type C struct {
	providers []Provider
}

// New is a constructor method which initializes configuration without providers.
func New() *C {
	return &C{
		providers: []Provider{},
	}
}

// Default is a constructor method which initializes default providers.
func Default() *C {
	return &C{
		providers: []Provider{
			&Yaml{},
			&Env{},
		},
	}
}

// WithProviders adds providers which will be used as configuration sources.
func (c *C) WithProviders(providers ...Provider) {
	c.providers = append(c.providers, providers...)
}

// Parse loops through providers and parses configuration.
func (c *C) Parse(config interface{}) error {
	cfgVal := reflect.ValueOf(config)

	err := validateConfig(cfgVal)
	if err != nil {
		return err
	}

	for _, p := range c.providers {
		source := reflect.New(reflect.TypeOf(config).Elem())
		err := p.Provide(source.Interface())
		if err != nil {
			return err
		}

		mergeConfig(source, cfgVal)
	}

	return nil
}

// Parse is a helper method which initializes internal configuration and parses the config.
func Parse(config interface{}) error {
	c := Default()
	return c.Parse(config)
}

func mergeConfig(source reflect.Value, target reflect.Value) {
	s := source.Elem()
	t := target.Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.CanSet() && !s.Field(i).IsZero() {
			f.Set(s.Field(i))
		}
	}
}

func validateConfig(cfgVal reflect.Value) error {
	if cfgVal.Kind() != reflect.Ptr {
		return errors.New("configuration must be a pointer")
	}

	if cfgVal.Elem().Kind() != reflect.Struct {
		return errors.New("value of configuration pointer must be a struct")
	}

	return nil
}
