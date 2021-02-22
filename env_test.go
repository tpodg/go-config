package config

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

const (
	stringField  = "string field val"
	intField     = 123456
	boolField    = true
	nestedString = "nested string field val"
	float32Field = 123.321
	nestedInt16  = -111
	uint8Filed   = 11
)

func TestFullEnvConfigWithPrefix(t *testing.T) {
	setUpEnv("PREF")

	e := &Env{Prefix: "PREF"}
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
	if cfg.NestedStruct.NestedString != nestedString {
		t.Errorf("Value is '%s', but %q expected", cfg.NestedStruct.NestedString, nestedString)
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

	e := &Env{}
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
}

func setUpEnv(prefix string) {
	p := ""
	if prefix != "" {
		p = prefix + "_"
	}
	_ = os.Setenv(p+"STRINGFIELD", stringField)
	_ = os.Setenv(p+"INTFIELD", strconv.Itoa(intField))
	_ = os.Setenv(p+"BOOLFIELD", strconv.FormatBool(boolField))
	_ = os.Setenv(p+"NESTEDSTRUCT_NESTEDSTRING", nestedString)
	_ = os.Setenv(p+"NESTEDSTRUCT_FLOAT32FIELD", fmt.Sprintf("%f", float32Field))
	_ = os.Setenv(p+"NESTEDSTRUCT_ANOTHERLEVEL_NESTEDINT16", strconv.Itoa(nestedInt16))
	_ = os.Setenv(p+"NESTEDSTRUCT_ANOTHERLEVEL_UINT8FIELD", strconv.Itoa(uint8Filed))
}
