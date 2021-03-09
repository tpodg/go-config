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
			if prefix != "" {
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
	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}
	val, ok := os.LookupEnv(strings.ToUpper(prefix) + strings.ToUpper(tField.Name))

	if ok && vField.CanSet() {
		switch vField.Kind() {
		case reflect.String:
			vField.SetString(val)
		case reflect.Bool:
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			vField.SetBool(bVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var iVal int64
			var err error
			if vField.Type().PkgPath() == "time" && vField.Type().Name() == "Duration" {
				var d time.Duration
				d, err = time.ParseDuration(val)
				iVal = int64(d)
			} else {
				iVal, err = strconv.ParseInt(val, 0, vField.Type().Bits())
			}
			if err != nil {
				return err
			}
			vField.SetInt(iVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			iVal, err := strconv.ParseUint(val, 0, vField.Type().Bits())
			if err != nil {
				return err
			}
			vField.SetUint(iVal)
		case reflect.Float32, reflect.Float64:
			fVal, err := strconv.ParseFloat(val, vField.Type().Bits())
			if err != nil {
				return err
			}
			vField.SetFloat(fVal)
		}
	}

	return nil
}
