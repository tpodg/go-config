package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYaml_Provide(t *testing.T) {
	dir, err := execDir()
	if err != nil {
		t.Fatal(err)
	}

	// Helper to write config file and handle cleanup
	writeConfig := func(t *testing.T, path string, content string) string {
		fullPath := path
		if !filepath.IsAbs(path) {
			fullPath = filepath.Join(dir, path)
		}
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(fullPath) })
		return path
	}

	tests := []struct {
		name        string
		path        string
		content     string
		expectedVal string
		expectErr   bool
	}{
		{
			name:        "Absolute path",
			path:        filepath.Join(t.TempDir(), "abs.yaml"),
			content:     "stringfield: absolute value",
			expectedVal: "absolute value",
		},
		{
			name:        "Relative path",
			path:        "test_relative.yaml",
			content:     "stringfield: relative value",
			expectedVal: "relative value",
		},
		{
			name:        "Default path",
			path:        "",
			content:     "stringfield: default value",
			expectedVal: "default value",
		},
		{
			name:      "Missing file",
			path:      "non_existent.yaml",
			expectErr: true,
		},
		{
			name:      "Invalid YAML",
			path:      "invalid.yaml",
			content:   "invalid: : :",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if tt.content != "" || tt.name == "Default path" {
				p := tt.path
				if p == "" {
					p = "config.yaml"
				}
				path = writeConfig(t, p, tt.content)
				if tt.path == "" {
					path = ""
				}
			}

			c := New()
			c.WithProviders(&Yaml{Path: path})

			cfg := struct {
				StringField string
			}{}

			err := c.Parse(&cfg)
			if (err != nil) != tt.expectErr {
				t.Fatalf("Parse() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr && cfg.StringField != tt.expectedVal {
				t.Errorf("Expected %q, got %q", tt.expectedVal, cfg.StringField)
			}
		})
	}
}
