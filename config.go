// Package config provides support for application configuration. Configuration values can be
// provided from multiple sources / providers. If the same value is configured in multiple
// sources, value from the source with the highest priority (added last to the slice of providers) will be applied.
package config

import (
	"errors"
	"reflect"
	"strings"
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
	mergeStructValue(source.Elem(), target.Elem())
}

func mergeMap(source reflect.Value, target reflect.Value) {
	if !target.CanSet() || source.IsZero() {
		return
	}
	if target.IsNil() {
		target.Set(reflect.MakeMapWithSize(target.Type(), source.Len()))
	}
	for _, key := range source.MapKeys() {
		mergedKey := key
		if key.Kind() == reflect.String {
			if tKey, ok := findStringMapKey(target, key.String()); ok {
				mergedKey = tKey
			}
		}
		sVal := source.MapIndex(key)
		tVal := target.MapIndex(mergedKey)
		if tVal.IsValid() {
			if merged, ok := mergeMapValue(sVal, tVal); ok {
				target.SetMapIndex(mergedKey, merged)
				continue
			}
		}
		target.SetMapIndex(mergedKey, sVal)
	}
}

func mergeMapValue(source reflect.Value, target reflect.Value) (reflect.Value, bool) {
	src := unwrapInterfaceValue(source)
	dst := unwrapInterfaceValue(target)
	if !src.IsValid() || !dst.IsValid() {
		return reflect.Value{}, false
	}
	if src.Kind() != reflect.Map || dst.Kind() != reflect.Map {
		return reflect.Value{}, false
	}
	if src.Type() != dst.Type() {
		return reflect.Value{}, false
	}

	merged := reflect.MakeMapWithSize(dst.Type(), dst.Len()+src.Len())
	for _, key := range dst.MapKeys() {
		merged.SetMapIndex(key, dst.MapIndex(key))
	}
	for _, key := range src.MapKeys() {
		sVal := src.MapIndex(key)
		if tVal := merged.MapIndex(key); tVal.IsValid() {
			if next, ok := mergeMapValue(sVal, tVal); ok {
				merged.SetMapIndex(key, next)
				continue
			}
		}
		merged.SetMapIndex(key, sVal)
	}
	return merged, true
}

func unwrapInterfaceValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

func mergeSlice(source reflect.Value, target reflect.Value) {
	if !target.CanSet() || source.IsNil() || source.Len() == 0 {
		return
	}
	if target.IsNil() {
		target.Set(reflect.MakeSlice(target.Type(), source.Len(), source.Len()))
	} else if target.Len() < source.Len() {
		s := reflect.MakeSlice(target.Type(), source.Len(), source.Len())
		reflect.Copy(s, target)
		target.Set(s)
	}

	for i := 0; i < source.Len(); i++ {
		sVal := source.Index(i)
		tVal := target.Index(i)
		mergeSliceElement(sVal, tVal)
	}
}

func mergeSliceElement(source reflect.Value, target reflect.Value) {
	if source.Kind() == reflect.Ptr {
		if source.IsNil() {
			return
		}
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		mergeValue(source.Elem(), target.Elem())
		return
	}
	mergeValue(source, target)
}

func mergeValue(source reflect.Value, target reflect.Value) {
	switch source.Kind() {
	case reflect.Struct:
		mergeStructValue(source, target)
	case reflect.Map:
		mergeMap(source, target)
	case reflect.Slice:
		mergeSlice(source, target)
	default:
		if !source.IsZero() && target.CanSet() {
			target.Set(source)
		}
	}
}

func mergeStructValue(source reflect.Value, target reflect.Value) {
	for i := 0; i < target.NumField(); i++ {
		f := target.Field(i)
		if !f.CanSet() {
			continue
		}
		s := source.Field(i)
		switch f.Kind() {
		case reflect.Struct:
			mergeStructValue(s, f)
		case reflect.Map:
			mergeMap(s, f)
		case reflect.Slice:
			mergeSlice(s, f)
		default:
			if !s.IsZero() {
				f.Set(s)
			}
		}
	}
}

func findStringMapKey(m reflect.Value, key string) (reflect.Value, bool) {
	keyVal := reflect.ValueOf(key)
	if keyVal.Type().AssignableTo(m.Type().Key()) {
		if val := m.MapIndex(keyVal); val.IsValid() {
			return keyVal, true
		}
	}
	keyLower := strings.ToLower(key)
	for _, k := range m.MapKeys() {
		if strings.ToLower(k.String()) == keyLower {
			return k, true
		}
	}
	return reflect.Value{}, false
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
