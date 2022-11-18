package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Env is a provider for configuration using environment variables
type Env struct {
	// Prefix of each environment variable used for configuration, no prefix will be used if not set
	Prefix string
}

// Provide loads configuration from environment variables
func (e *Env) Provide(config interface{}) error {
	err := provide(e.Prefix, reflect.ValueOf(config))
	if err != nil {
		return err
	}

	return nil
}

func provide(prefix string, config reflect.Value) error {
	cfgVal := config.Elem()
	tt := cfgVal.Type()

	for i := 0; i < cfgVal.NumField(); i++ {
		vf := cfgVal.Field(i)
		tf := tt.Field(i)

		if vf.Kind() == reflect.Struct {
			if prefix != "" && !strings.HasSuffix(prefix, "_") {
				prefix += "_"
			}
			err := provide(prefix+tf.Name, vf.Addr())
			if err != nil {
				return err
			}
		} else {
			err := parseValue(prefix, vf, tf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseValue(prefix string, vField reflect.Value, tField reflect.StructField) error {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix = strings.ToUpper(prefix) + "_"
	}

	envVal, ok := os.LookupEnv(strings.ToUpper(prefix) + strings.ToUpper(tField.Name))
	if ok && envVal != "" && vField.CanSet() {
		if err := processField(vField, envVal); err != nil {
			return err
		}
	}
	return nil
}

func processField(vField reflect.Value, envVal string) error {
	switch vField.Kind() {
	case reflect.String:
		vField.SetString(envVal)
	case reflect.Bool:
		val, err := strconv.ParseBool(envVal)
		if err != nil {
			return err
		}
		vField.SetBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var val int64
		var err error
		if vField.Type().PkgPath() == "time" && vField.Type().Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(envVal)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(envVal, 0, vField.Type().Bits())
		}
		if err != nil {
			return err
		}
		vField.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(envVal, 0, vField.Type().Bits())
		if err != nil {
			return err
		}
		vField.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(envVal, vField.Type().Bits())
		if err != nil {
			return err
		}
		vField.SetFloat(val)
	case reflect.Slice:
		vals := strings.Split(envVal, ",")
		s := reflect.MakeSlice(vField.Type(), len(vals), len(vals))
		for i, val := range vals {
			if err := processField(s.Index(i), strings.TrimSpace(val)); err != nil {
				return err
			}
		}
		vField.Set(s)
	}
	return nil
}
