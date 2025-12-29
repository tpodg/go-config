package config

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	stringField  = "string field val"
	intField     = 123456
	boolField    = true
	durField     = 10 * time.Second
	nestedString = "nested string field val"
	float32Field = 123.321
	nestedInt16  = -111
	uint8Filed   = 11
)

func TestFullEnvConfigWithPrefix(t *testing.T) {
	setUpEnv("PREF")

	e := Env{Prefix: "PREF"}
	cfg := testCfg{}
	err := e.Provide(&cfg)
	if err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}
	if cfg.StringField != stringField {
		t.Errorf("Value is '%s', but %q expected", cfg.StringField, stringField)
	}
	if cfg.IntField != intField {
		t.Errorf("Value is '%d', but %d expected", cfg.IntField, intField)
	}
	if cfg.BoolField != boolField {
		t.Errorf("Value is '%t', but %t expected", cfg.BoolField, boolField)
	}
	if cfg.DurField != durField {
		t.Errorf("Value is '%v', but %v expected", cfg.DurField, durField)
	}
	if cfg.NestedStruct.StringSlice[0] != nestedString {
		t.Errorf("Value is '%s', but %q expected", cfg.NestedStruct.StringSlice[0], nestedString)
	}
	if cfg.NestedStruct.Float32Field != float32Field {
		t.Errorf("Value is '%f', but %f expected", cfg.NestedStruct.Float32Field, float32Field)
	}
	if cfg.NestedStruct.AnotherLevel.NestedInt16 != nestedInt16 {
		t.Errorf("Value is '%d', but %d expected", cfg.NestedStruct.AnotherLevel.NestedInt16, nestedInt16)
	}
	if cfg.NestedStruct.AnotherLevel.Uint8Field != uint8Filed {
		t.Errorf("Value is '%d', but %d expected", cfg.NestedStruct.AnotherLevel.Uint8Field, uint8Filed)
	}
}

func TestEnvConfigWithoutPrefix(t *testing.T) {
	setUpEnv("")

	e := Env{}
	cfg := testCfg{}
	err := e.Provide(&cfg)
	if err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}
	if cfg.StringField != stringField {
		t.Errorf("Value is '%s', but %q expected", cfg.StringField, stringField)
	}
	if cfg.NestedStruct.AnotherLevel.Uint8Field != uint8Filed {
		t.Errorf("Value is '%d', but %d expected", cfg.NestedStruct.AnotherLevel.Uint8Field, uint8Filed)
	}
	if cfg.SecondNestedStruct.StringField != nestedString {
		t.Errorf("Value is '%s', but %q expected", cfg.SecondNestedStruct.StringField, nestedString)
	}
	if cfg.SecondNestedStruct.SecondStringField != nestedString {
		t.Errorf("Value is '%s', but %q expected", cfg.SecondNestedStruct.SecondStringField, nestedString)
	}
	if cfg.SecondNestedStruct.StructField.StringField != nestedString {
		t.Errorf("Value is '%s', but %q expected", cfg.SecondNestedStruct.StructField.StringField, nestedString)
	}
}

func TestEnvConfigEmptyString(t *testing.T) {
	_ = os.Setenv("STRINGFIELD", "")
	_ = os.Setenv("INTFIELD", "")

	e := Env{}
	cfg := testCfg{
		StringField: stringField,
		IntField:    intField,
	}

	err := e.Provide(&cfg)
	if err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}
	if cfg.StringField != stringField {
		t.Errorf("Value is '%s', but %q expected", cfg.StringField, stringField)
	}
	if cfg.IntField != intField {
		t.Errorf("Value is '%d', but %d expected", cfg.IntField, intField)
	}
}

func TestEnvConfigSlice(t *testing.T) {
	_ = os.Setenv("NESTEDSTRUCT_STRINGSLICE_0", "val 1")
	_ = os.Setenv("NESTEDSTRUCT_STRINGSLICE_1", "val   2")
	_ = os.Setenv("NESTEDSTRUCT_STRINGSLICE_2", "val3")

	cfg := testCfg{}

	e := Env{}
	if err := e.Provide(&cfg); err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}

	exp := "val 1"
	if cfg.NestedStruct.StringSlice[0] != exp {
		t.Errorf("Value is '%s', but %q expected", cfg.NestedStruct.StringSlice[0], exp)
	}
	exp = "val   2"
	if cfg.NestedStruct.StringSlice[1] != exp {
		t.Errorf("Value is '%s', but %q expected", cfg.NestedStruct.StringSlice[1], exp)
	}
	exp = "val3"
	if cfg.NestedStruct.StringSlice[2] != exp {
		t.Errorf("Value is '%s', but %q expected", cfg.NestedStruct.StringSlice[2], exp)
	}
}

func TestNestedStructs(t *testing.T) {
	_ = os.Setenv("WRAPPER_NESTED1_NESTED1VAL", "val 1")
	_ = os.Setenv("WRAPPER_NESTED2_NESTED2VAL", "val 2")

	var cfg struct {
		Wrapper struct {
			Nested1 struct {
				Nested1Val string
			}
			Nested2 struct {
				Nested2Val string
			}
		}
	}

	e := Env{}
	if err := e.Provide(&cfg); err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}

	if cfg.Wrapper.Nested1.Nested1Val != "val 1" {
		t.Errorf("Value is '%s', but %q expected", cfg.Wrapper.Nested1.Nested1Val, "val 1")
	}
	if cfg.Wrapper.Nested2.Nested2Val != "val 2" {
		t.Errorf("Value is '%s', but %q expected", cfg.Wrapper.Nested2.Nested2Val, "val 2")
	}
}

func TestEnvConfigYamlTags(t *testing.T) {
	_ = os.Setenv("WRAPPER_TAG_ONE", "val-1")
	_ = os.Setenv("WRAPPER_TAG_TWO", "val-2")

	var cfg struct {
		Wrapper struct {
			TagOne string `yaml:"tag_one"`
			TagTwo string `yaml:"tag_two"`
		} `yaml:"wrapper"`
	}

	e := Env{}
	if err := e.Provide(&cfg); err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}

	if cfg.Wrapper.TagOne != "val-1" {
		t.Errorf("Value is '%s', but %q expected", cfg.Wrapper.TagOne, "val-1")
	}
	if cfg.Wrapper.TagTwo != "val-2" {
		t.Errorf("Value is '%s', but %q expected", cfg.Wrapper.TagTwo, "val-2")
	}
}

func TestEnvConfigMap(t *testing.T) {
	_ = os.Setenv("MAPFIELD_BUILD", "go build")
	_ = os.Setenv("MAPFIELD_TEST", "go test")

	var cfg struct {
		MapField map[string]any
	}

	e := Env{}
	if err := e.Provide(&cfg); err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}

	if len(cfg.MapField) != 2 {
		t.Fatalf("Expected 2 map entries, got %d", len(cfg.MapField))
	}
	if cfg.MapField["build"] != "go build" {
		t.Errorf("Value is '%v', but %q expected", cfg.MapField["build"], "go build")
	}
}

func TestEnvConfigPointer(t *testing.T) {
	_ = os.Setenv("USE_AGENT", "true")

	var cfg struct {
		UseAgent *bool `yaml:"use_agent"`
	}

	e := Env{}
	if err := e.Provide(&cfg); err != nil {
		t.Fatalf("No error expected, but was: %v\n", err)
	}

	if cfg.UseAgent == nil {
		t.Fatal("UseAgent should not be nil")
	}
	if *cfg.UseAgent != true {
		t.Errorf("Value is '%t', but %t expected", *cfg.UseAgent, true)
	}
}

func setUpEnv(prefix string) {
	p := ""
	if prefix != "" {
		p = prefix + "_"
	}
	_ = os.Setenv(p+"STRINGFIELD", stringField)
	_ = os.Setenv(p+"INTFIELD", strconv.Itoa(intField))
	_ = os.Setenv(p+"BOOLFIELD", strconv.FormatBool(boolField))
	_ = os.Setenv(p+"DURFIELD", "10s")
	_ = os.Setenv(p+"NESTEDSTRUCT_STRINGSLICE_0", nestedString)
	_ = os.Setenv(p+"NESTEDSTRUCT_FLOAT32FIELD", fmt.Sprintf("%f", float32Field))
	_ = os.Setenv(p+"NESTEDSTRUCT_ANOTHERLEVEL_NESTEDINT16", strconv.Itoa(nestedInt16))
	_ = os.Setenv(p+"NESTEDSTRUCT_ANOTHERLEVEL_UINT8FIELD", strconv.Itoa(uint8Filed))
	_ = os.Setenv(p+"SECONDNESTEDSTRUCT_STRINGFIELD", nestedString)
	_ = os.Setenv(p+"SECONDNESTEDSTRUCT_STRUCTFIELD_STRINGFIELD", nestedString)
	_ = os.Setenv(p+"SECONDNESTEDSTRUCT_SECONDSTRINGFIELD", nestedString)
}
