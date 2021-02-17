package config

import (
	"reflect"
	"strings"
	"testing"
)

// Count of default providers
var provCount = 2

// Additional providers used in tests
type pFull struct{}
type pSimple struct {
	priority int
}

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

	if !strings.Contains(err.Error(), "configuration struct must be a pointer") {
		t.Errorf("Interface kind check should fail, but was: %v", err)
	}
}

func TestUniquePriority(t *testing.T) {
	c := New()
	c.WithProviders(&Env{})
	c.WithProviders(&pSimple{priority: 30})
	err := c.Parse(&testCfg{})
	if err == nil {
		t.Fatalf("Error expected, but there is none.")
	}

	if !strings.Contains(err.Error(), "get priorities must be unique") {
		t.Errorf("Uniqueness check should fail, but was: %v", err)
	}
}

func TestPriorityBelowOne(t *testing.T) {
	c := New()
	c.WithProviders(&pSimple{priority: 0})
	err := c.Parse(&testCfg{})
	if err == nil {
		t.Fatalf("Error expected, but there is none.")
	}

	if !strings.Contains(err.Error(), "priority below 1 is not allowed") {
		t.Errorf("Priority check should fail, but was: %v", err)
	}
}

func TestParseConfig(t *testing.T) {
	c := New()
	c.WithProviders(&pFull{})

	conf := &testCfg{}
	err := c.Parse(conf)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if conf.StringField != "1234string" {
		t.Errorf("Value is '%s', but '1234string' expected", conf.StringField)
	}
	if conf.IntField != 123 {
		t.Errorf("Value is '%d', but '123' expected", conf.IntField)
	}
	if conf.NestedStruct.NestedString != "nested value" {
		t.Errorf("Value is '%s', but 'nested value' expected", conf.NestedStruct.NestedString)
	}
	if conf.NestedStruct.AnotherLevel.NestedInt != 321 {
		t.Errorf("Value is '%d', but '321' expected", conf.NestedStruct.AnotherLevel.NestedInt)
	}
}

func TestPriorityAndSkipNonDefinedValue(t *testing.T) {
	c := New()
	c.WithProviders(&pFull{}, &pSimple{priority: 90})

	conf := &testCfg{}
	err := c.Parse(conf)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	if conf.StringField != "String from simple" {
		t.Errorf("Value is '%s', but 'String from simple' expected", conf.StringField)
	}
	if conf.IntField != 9999 {
		t.Errorf("Value is '%d', but '9999' expected", conf.IntField)
	}
	if conf.NestedStruct.NestedString != "nested value" {
		t.Errorf("Value is '%s', but 'nested value' expected", conf.NestedStruct.NestedString)
	}
	if conf.NestedStruct.AnotherLevel.NestedInt != 321 {
		t.Errorf("Value is '%d', but '321' expected", conf.NestedStruct.AnotherLevel.NestedInt)
	}
}

func assertProviderCount(t *testing.T, expected int, actual int) {
	if actual != expected {
		t.Fatalf("Configured providers: %d, but %d expected", actual, expected)
	}
}

func (p *pFull) Priority() int {
	return 100
}

func (p *pSimple) Priority() int {
	return p.priority
}

type testCfg struct {
	StringField  string
	IntField     int
	NestedStruct struct {
		NestedString string
		AnotherLevel struct {
			NestedInt int
		}
	}
}

func (p *pFull) Provide(config interface{}) error {
	cfg := &testCfg{
		StringField: "1234string",
		IntField:    123,
		NestedStruct: struct {
			NestedString string
			AnotherLevel struct {
				NestedInt int
			}
		}{
			NestedString: "nested value",
			AnotherLevel: struct {
				NestedInt int
			}{
				NestedInt: 321,
			},
		},
	}

	parse(config, cfg)

	return nil
}

func (p *pSimple) Provide(config interface{}) error {
	cfg := &testCfg{
		StringField: "String from simple",
		IntField:    9999,
	}

	parse(config, cfg)

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
