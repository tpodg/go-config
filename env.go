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

	env := readEnvVars()
	return applyEnvOverrides(e.Prefix, reflect.ValueOf(config), env)
}

func provide(prefix string, config reflect.Value) error {
	cfgVal := config.Elem()
	tt := cfgVal.Type()

	for i := 0; i < cfgVal.NumField(); i++ {
		vf := cfgVal.Field(i)
		tf := tt.Field(i)
		fieldName, ok := envFieldName(tf)
		if !ok {
			continue
		}

		if vf.Kind() == reflect.Struct {
			nextPrefix := prefix
			if nextPrefix != "" && !strings.HasSuffix(nextPrefix, "_") {
				nextPrefix += "_"
			}
			err := provide(nextPrefix+fieldName, vf.Addr())
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
	fieldName, ok := envFieldName(tField)
	if !ok {
		return nil
	}

	if vField.Kind() == reflect.Slice || vField.Kind() == reflect.Map {
		return nil
	}

	envPrefix := strings.ToUpper(prefix)
	if envPrefix != "" && !strings.HasSuffix(envPrefix, "_") {
		envPrefix += "_"
	}

	envVal, ok := os.LookupEnv(envPrefix + strings.ToUpper(fieldName))
	if ok && envVal != "" && vField.CanSet() {
		if err := processField(vField, envVal); err != nil {
			return err
		}
	}
	return nil
}

func processField(vField reflect.Value, envVal string) error {
	if vField.Kind() == reflect.Ptr {
		if vField.IsNil() {
			vField.Set(reflect.New(vField.Type().Elem()))
		}
		return processField(vField.Elem(), envVal)
	}

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
	case reflect.Interface:
		vField.Set(reflect.ValueOf(envVal))
	}
	return nil
}

func isComplexType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Interface:
		return true
	default:
		return false
	}
}

func envFieldName(tField reflect.StructField) (string, bool) {
	tag := tField.Tag.Get("yaml")
	if tag != "" {
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			return "", false
		}
		if name != "" {
			return name, true
		}
	}

	return tField.Name, true
}

type envVars struct {
	values map[string]string
	keys   []string
}

func readEnvVars() envVars {
	env := envVars{
		values: make(map[string]string),
	}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToUpper(parts[0])
		if _, ok := env.values[key]; !ok {
			env.keys = append(env.keys, key)
		}
		env.values[key] = parts[1]
	}
	return env
}

func applyEnvOverrides(prefix string, config reflect.Value, env envVars) error {
	if config.Kind() != reflect.Ptr || config.IsNil() {
		return nil
	}
	cfg := config.Elem()
	if cfg.Kind() != reflect.Struct {
		return nil
	}
	return applyOverrides(prefix, cfg, false, env)
}

func applyOverrides(prefix string, v reflect.Value, setScalars bool, env envVars) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)
		fieldName, ok := envFieldName(tf)
		if !ok {
			continue
		}

		switch vf.Kind() {
		case reflect.Struct:
			nextPrefix := joinPrefix(prefix, fieldName)
			if err := applyOverrides(nextPrefix, vf, setScalars, env); err != nil {
				return err
			}
		case reflect.Slice:
			if err := applySliceOverrides(prefix, fieldName, vf, tf, setScalars, env); err != nil {
				return err
			}
		case reflect.Map:
			if err := applyMapOverrides(prefix, fieldName, vf, tf, setScalars, env); err != nil {
				return err
			}
		default:
			if setScalars {
				if err := parseValue(prefix, vf, tf); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func applySliceOverrides(prefix, fieldName string, vField reflect.Value, tField reflect.StructField, setScalars bool, env envVars) error {
	if !vField.CanSet() {
		return nil
	}

	base := joinPrefix(prefix, fieldName)
	return applySliceValueOverrides(base, vField, env, false)
}

func collectSliceIndices(prefix string, keys []string) map[int]struct{} {
	indices := map[int]struct{}{}
	for _, key := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		rest := key[len(prefix):]
		if rest == "" {
			continue
		}
		idxStr := rest
		if pos := strings.IndexByte(rest, '_'); pos != -1 {
			idxStr = rest[:pos]
		}
		if idxStr == "" {
			continue
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 {
			continue
		}
		indices[idx] = struct{}{}
	}
	return indices
}

func applyMapOverrides(prefix, fieldName string, vField reflect.Value, tField reflect.StructField, setScalars bool, env envVars) error {
	if !vField.CanSet() {
		return nil
	}
	base := joinPrefix(prefix, fieldName)
	return applyMapValueOverrides(base, vField, env, false)
}

func applySliceValueOverrides(base string, vField reflect.Value, env envVars, setDirect bool) error {
	if !vField.CanSet() {
		return nil
	}

	baseUpper := strings.ToUpper(base)
	if setDirect {
		if val, ok := env.values[baseUpper]; ok && val != "" {
			if err := processField(vField, val); err != nil {
				return err
			}
		}
	}

	idxPrefix := baseUpper
	if !strings.HasSuffix(idxPrefix, "_") {
		idxPrefix += "_"
	}

	indices := collectSliceIndices(idxPrefix, env.keys)
	if len(indices) == 0 {
		return nil
	}

	maxIdx := -1
	for idx := range indices {
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	if maxIdx < 0 {
		return nil
	}

	if vField.IsNil() || vField.Len() <= maxIdx {
		newLen := maxIdx + 1
		s := reflect.MakeSlice(vField.Type(), newLen, newLen)
		reflect.Copy(s, vField)
		vField.Set(s)
	}

	for idx := range indices {
		elem := vField.Index(idx)
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}

		idxKey := idxPrefix + strconv.Itoa(idx)
		if val, ok := env.values[idxKey]; ok && val != "" {
			if err := processField(elem, val); err != nil {
				return err
			}
		}

		if err := applyValueOverrides(idxKey, elem, env); err != nil {
			return err
		}
	}

	return nil
}

func applyMapValueOverrides(base string, vField reflect.Value, env envVars, setDirect bool) error {
	if !vField.CanSet() {
		return nil
	}

	baseUpper := strings.ToUpper(base)
	if setDirect {
		if val, ok := env.values[baseUpper]; ok && val != "" {
			if err := processField(vField, val); err != nil {
				return err
			}
		}
	}

	keyPrefix := baseUpper
	if !strings.HasSuffix(keyPrefix, "_") {
		keyPrefix += "_"
	}

	keys := collectMapKeys(keyPrefix, vField.Type().Elem(), env.keys)
	if len(keys) == 0 {
		return nil
	}

	if vField.IsNil() {
		vField.Set(reflect.MakeMapWithSize(vField.Type(), len(keys)))
	}

	for keyUpper := range keys {
		keyVal, err := parseMapKey(keyUpper, vField.Type().Key())
		if err != nil {
			return err
		}
		mapKey := keyVal
		if vField.Type().Key().Kind() == reflect.String {
			if existing, ok := findStringMapKey(vField, keyVal.String()); ok {
				mapKey = existing
			}
		}

		elemVal := vField.MapIndex(mapKey)
		if !elemVal.IsValid() {
			elemVal = reflect.New(vField.Type().Elem()).Elem()
		} else if elemVal.Kind() == reflect.Ptr && elemVal.IsNil() {
			elemVal = reflect.New(elemVal.Type().Elem())
		}

		entryPrefix := keyPrefix + keyUpper
		if val, ok := env.values[entryPrefix]; ok && val != "" {
			if err := processField(elemVal, val); err != nil {
				return err
			}
		}

		if err := applyValueOverrides(entryPrefix, elemVal, env); err != nil {
			return err
		}

		vField.SetMapIndex(mapKey, elemVal)
	}

	return nil
}

func applyValueOverrides(prefix string, vField reflect.Value, env envVars) error {
	if vField.Kind() == reflect.Ptr {
		if vField.IsNil() {
			vField.Set(reflect.New(vField.Type().Elem()))
		}
		return applyValueOverrides(prefix, vField.Elem(), env)
	}
	switch vField.Kind() {
	case reflect.Struct:
		return applyOverrides(prefix, vField, true, env)
	case reflect.Slice:
		return applySliceValueOverrides(prefix, vField, env, false)
	case reflect.Map:
		return applyMapValueOverrides(prefix, vField, env, false)
	default:
		return nil
	}
}

func collectMapKeys(prefix string, elemType reflect.Type, keys []string) map[string]struct{} {
	keySet := map[string]struct{}{}
	elemComplex := isComplexType(elemType)
	if elemType.Kind() == reflect.Interface {
		elemComplex = false
	}
	for _, key := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		rest := key[len(prefix):]
		if rest == "" {
			continue
		}
		keyPart := rest
		if elemComplex {
			if pos := strings.IndexByte(rest, '_'); pos != -1 {
				keyPart = rest[:pos]
			}
		}
		if keyPart == "" {
			continue
		}
		keySet[keyPart] = struct{}{}
	}
	return keySet
}

func parseMapKey(key string, keyType reflect.Type) (reflect.Value, error) {
	k := reflect.New(keyType).Elem()
	if keyType.Kind() == reflect.String {
		k.SetString(strings.ToLower(key))
		return k, nil
	}
	switch keyType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
	default:
		return reflect.Value{}, strconv.ErrSyntax
	}
	if err := processField(k, key); err != nil {
		return reflect.Value{}, err
	}
	return k, nil
}

func joinPrefix(prefix, fieldName string) string {
	if prefix == "" {
		return fieldName
	}
	if strings.HasSuffix(prefix, "_") {
		return prefix + fieldName
	}
	return prefix + "_" + fieldName
}
