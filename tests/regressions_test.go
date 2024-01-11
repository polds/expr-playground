package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/polds/expr-playground/eval"
)

// TestExamples serves as a regression test for the examples presented in the playground UI.
// If any of the examples change, this test will fail, to help ensure the playground UI is
// updated accordingly, and especially so we don't accidentally push a broken sample.
func TestExamples(t *testing.T) {
	examples := setup(t)

	// lookup should exactly match the "name" field in the examples.yaml file.
	tests := []struct {
		lookup  string
		want    string
		wantErr bool
	}{
		{
			lookup: "default",
			want:   "true",
		},
		{
			lookup: "Check image registry",
			want:   "true",
		},
		{
			lookup: "Disallow HostPorts",
			want:   "false",
		},
		{
			lookup: "Require non-root containers",
			want:   "false",
		},
		{
			lookup: "Drop ALL capabilities",
			want:   "true",
		},
		{
			lookup: "Semantic version check for image tags (Regex)",
			want:   "false",
		},
		{
			lookup:  "URLs",
			wantErr: true,
		},
		{
			lookup: "Check JWT custom claims",
			want:   "true",
		},
		{
			lookup: "Optional",
			want:   "fallback",
		},
		{
			lookup: "Duration and timestamp",
			want:   "true",
		},
		{
			lookup:  "Quantity",
			wantErr: true,
		},
		{
			lookup: "Access Log Filtering",
			want:   "true",
		},
		{
			lookup: "Custom Metrics",
			want:   "echo",
		},
		{
			lookup:  "Blank",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.lookup, func(t *testing.T) {
			var exp Example
			for _, e := range examples {
				if e.Name == tc.lookup {
					exp = e
					break
				}
			}
			if exp.Name == "" {
				t.Fatalf("failed to find example %q", tc.lookup)
			}

			got, err := eval.Eval(exp.Expr, marshal(t, exp.Data))
			if (err != nil) != tc.wantErr {
				t.Errorf("Eval() got error %v, expected error %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			var obj map[string]AlwaysString
			if err := json.Unmarshal([]byte(got), &obj); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}
			if s := obj["result"].Value; s != tc.want {
				t.Errorf("Eval() got %q, expected %q", s, tc.want)
			}
		})
	}
	// Ensure these tests are updated when the examples are updated.
	// Not a perfect solution, but it's better than nothing.
	if len(examples) != len(tests) {
		t.Errorf("Regression test counts got %d, expected %d", len(tests), len(examples))
	}
}

type Example struct {
	Name string `yaml:"name"`
	Expr string `yaml:"expr"`
	Data string `yaml:"data"`
}

func setup(t *testing.T) []Example {
	t.Helper()

	out, err := os.ReadFile("../examples.yaml")
	if err != nil {
		t.Fatalf("failed to read examples.yaml: %v", err)
	}
	var examples struct {
		Examples []Example `yaml:"examples"`
	}
	if err := yaml.Unmarshal(out, &examples); err != nil {
		t.Fatalf("failed to unmarshal examples.yaml: %v", err)
	}
	return examples.Examples
}

// Attempt to get the data into either yaml or json format.
func marshal(t *testing.T, s string) map[string]any {
	t.Helper()

	var v map[string]any
	if yamlErr := yaml.Unmarshal([]byte(s), &v); yamlErr != nil {
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Errorf("failed to unmarshal %q as yaml: %v", s, yamlErr)
			t.Fatalf("failed to unmarshal %q as json: %v", s, err)
		}
	}
	return v
}

// AlwaysString attempts to unmarshal the value as a string.
type AlwaysString struct {
	Value string
}

func (c *AlwaysString) UnmarshalJSON(b []byte) error {
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case bool:
		c.Value = strconv.FormatBool(v)
	case string:
		c.Value = v
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
	return nil
}
