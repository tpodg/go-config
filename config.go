// Package config provides support for application configuration. Configuration values can be
// provided from multiple sources based on priority of specific configuration source (provider).
// If the same value is configured in multiple sources, value from the source with the highest
// priority will be applied.
package config

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

// Provider interface offers support for multiple configuration sources.
type Provider interface {
	// Priority of the configuration source. Priority 1 is considered the highest,
	// priorities below 1 are not allowed.
	Priority() int

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

// Parse loops through providers based on priority and parses configuration.
func (c *C) Parse(config interface{}) error {
	tVal := reflect.ValueOf(config)
	if tVal.Kind() != reflect.Ptr {
		return errors.New("configuration struct must be a pointer")
	}

	providers, err := c.providerMap()
	if err != nil {
		return err
	}

	priorities := make([]int, 0, len(providers))
	for k := range providers {
		priorities = append(priorities, k)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(priorities)))

	for _, p := range priorities {
		provider, ok := providers[p]
		if !ok {
			return errors.New(fmt.Sprintf("no provider with priority %d found", p))
		}

		source := reflect.New(reflect.TypeOf(config).Elem())
		err := provider.Provide(source.Interface())
		if err != nil {
			return err
		}

		mergeConfig(source, tVal)
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

func (c *C) providerMap() (map[int]Provider, error) {
	m := make(map[int]Provider, len(c.providers))

	for _, p := range c.providers {
		if p.Priority() < 1 {
			return nil, errors.New("priority below 1 is not allowed")
		}
		if m[p.Priority()] != nil {
			return nil, errors.New(fmt.Sprintf("get priorities must be unique: %T <-> %T", m[p.Priority()], p))
		}
		m[p.Priority()] = p
	}

	return m, nil
}
