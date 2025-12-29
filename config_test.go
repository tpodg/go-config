package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Count of default providers
var provCount = 2

// Additional providers used in tests
type pFull struct{}
type pSimple struct{}
type pMapBase struct{}
type pMapOverride struct{}

func TestInitializeNewConfigWithCustomProvider(t *testing.T) {
	c := New()
	assertProviderCount(t, 0, len(c.providers))

	c.WithProviders(&pFull{})
	assertProviderCount(t, 1, len(c.providers))
}

func TestInitializeNewConfigWithInternalAndCustomProvider(t *testing.T) {
	c := New()
	assertProviderCount(t, 0, len(c.providers))

	c.WithProviders(&Yaml{}, &Env{})
	assertProviderCount(t, 2, len(c.providers))

	c.WithProviders(&pSimple{})
	assertProviderCount(t, 3, len(c.providers))
}

func TestInitializeInternalConfigWithAdditionalProvider(t *testing.T) {
	c := Default()
	assertProviderCount(t, provCount, len(c.providers))

	c.WithProviders(&pFull{})
	assertProviderCount(t, provCount+1, len(c.providers))
}

func TestInterfaceKind(t *testing.T) {
	err := Parse(testCfg{})
	if err == nil {
		t.Fatalf("Error expected, but there is none.")
	}

	if !strings.Contains(err.Error(), "configuration must be a pointer") {
		t.Errorf("Interface kind check should fail, but was: %v", err)
	}
}

func TestInterfaceValueKind(t *testing.T) {
	config := "string"
	err := Parse(&config)
	if err == nil {
		t.Fatalf("Error expected, but there is none.")
	}

	if !strings.Contains(err.Error(), "value of configuration pointer must be a struct") {
		t.Errorf("Interface Value kind check should fail, but was: %v", err)
	}
}

func TestParseConfig(t *testing.T) {
	c := New()
	c.WithProviders(&pFull{})

	conf := testCfg{}
	err := c.Parse(&conf)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if conf.StringField != "1234string" {
		t.Errorf("Value is '%s', but '1234string' expected", conf.StringField)
	}
	if conf.IntField != 123 {
		t.Errorf("Value is '%d', but '123' expected", conf.IntField)
	}
	if conf.NestedStruct.StringSlice[1] != "val 2" {
		t.Errorf("Value is '%s', but 'val 2' expected", conf.NestedStruct.StringSlice[1])
	}
	if conf.NestedStruct.AnotherLevel.NestedInt16 != 321 {
		t.Errorf("Value is '%d', but '321' expected", conf.NestedStruct.AnotherLevel.NestedInt16)
	}
}

func TestSkipNonDefinedValue(t *testing.T) {
	c := New()
	c.WithProviders(&pFull{}, &pSimple{})

	conf := testCfg{}
	err := c.Parse(&conf)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if conf.StringField != "String from simple" {
		t.Errorf("Value is '%s', but 'String from simple' expected", conf.StringField)
	}
	if conf.IntField != 9999 {
		t.Errorf("Value is '%d', but '9999' expected", conf.IntField)
	}
	if conf.NestedStruct.StringSlice[1] != "val 2" {
		t.Errorf("Value is '%s', but 'val 2' expected", conf.NestedStruct.StringSlice[1])
	}
	if conf.NestedStruct.Float32Field != 4545.1212 {
		t.Errorf("Value is '%f', but '4545.1212' expected", conf.NestedStruct.Float32Field)
	}
	if conf.NestedStruct.AnotherLevel.NestedInt16 != 321 {
		t.Errorf("Value is '%d', but '321' expected", conf.NestedStruct.AnotherLevel.NestedInt16)
	}
}

func TestMergeMapConfig(t *testing.T) {
	c := New()
	c.WithProviders(&pMapBase{}, &pMapOverride{})

	cfg := mapCfg{}
	if err := c.Parse(&cfg); err != nil {
		t.Fatalf("%v\n", err)
	}

	if cfg.MapField["build"] != "go build env" {
		t.Errorf("Value is '%v', but %q expected", cfg.MapField["build"], "go build env")
	}
	if cfg.MapField["test"] != "go test" {
		t.Errorf("Value is '%v', but %q expected", cfg.MapField["test"], "go test")
	}

	nested, ok := cfg.MapField["nested"].(map[string]any)
	if !ok {
		t.Fatalf("Expected nested map, got %T", cfg.MapField["nested"])
	}
	if nested["a"] != 1 {
		t.Errorf("Value is '%v', but %d expected", nested["a"], 1)
	}
	if nested["b"] != 3 {
		t.Errorf("Value is '%v', but %d expected", nested["b"], 3)
	}
	if nested["c"] != 4 {
		t.Errorf("Value is '%v', but %d expected", nested["c"], 4)
	}
}

func assertProviderCount(t *testing.T, expected int, actual int) {
	if actual != expected {
		t.Fatalf("Configured providers: %d, but %d expected", actual, expected)
	}
}

type testCfg struct {
	StringField  string
	IntField     int
	BoolField    bool
	DurField     time.Duration
	NestedStruct struct {
		StringSlice  []string
		Float32Field float32
		AnotherLevel struct {
			NestedInt16 int16
			Uint8Field  uint8
		}
	}
	SecondNestedStruct struct {
		StringField string
		StructField struct {
			StringField string
		}
		SecondStringField string
	}
}

type mapCfg struct {
	MapField map[string]any
}

func (p *pFull) Provide(config interface{}) error {
	cfg := testCfg{
		StringField: "1234string",
		IntField:    123,
		NestedStruct: struct {
			StringSlice  []string
			Float32Field float32
			AnotherLevel struct {
				NestedInt16 int16
				Uint8Field  uint8
			}
		}{
			StringSlice: []string{"val 1", "val 2", "val 3"},
			AnotherLevel: struct {
				NestedInt16 int16
				Uint8Field  uint8
			}{
				NestedInt16: 321,
			},
		},
		SecondNestedStruct: struct {
			StringField string
			StructField struct {
				StringField string
			}
			SecondStringField string
		}{
			StringField: "nested-123",
			StructField: struct {
				StringField string
			}{
				StringField: "inner-nested",
			},
			SecondStringField: "nested-123",
		},
	}

	parse(config, &cfg)

	return nil
}

func (p *pSimple) Provide(config interface{}) error {
	cfg := testCfg{
		StringField: "String from simple",
		IntField:    9999,
		NestedStruct: struct {
			StringSlice  []string
			Float32Field float32
			AnotherLevel struct {
				NestedInt16 int16
				Uint8Field  uint8
			}
		}{
			Float32Field: 4545.1212,
		},
	}

	parse(config, &cfg)

	return nil
}

func (p *pMapBase) Provide(config interface{}) error {
	cfg := config.(*mapCfg)
	cfg.MapField = map[string]any{
		"build": "go build",
		"test":  "go test",
		"nested": map[string]any{
			"a": 1,
			"b": 2,
		},
	}
	return nil
}

func (p *pMapOverride) Provide(config interface{}) error {
	cfg := config.(*mapCfg)
	cfg.MapField = map[string]any{
		"build": "go build env",
		"nested": map[string]any{
			"b": 3,
			"c": 4,
		},
	}
	return nil
}

func parse(config interface{}, cfg *testCfg) {
	v := reflect.ValueOf(config).Elem()
	vn := reflect.ValueOf(cfg).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.CanSet() {
			f.Set(vn.Field(i))
		}
	}
}
