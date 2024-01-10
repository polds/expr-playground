// Copyright 2023 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eval

import (
	"encoding/json"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/google/go-cmp/cmp"
)

var input = map[string]any{
	"object": map[string]any{
		"replicas": 2,
		"href":     "https://user:pass@example.com:80/path?query=val#fragment",
		"image":    "registry.com/image:v0.0.0",
		"items":    []int{1, 2, 3},
		"abc":      []string{"a", "b", "c"},
		"memory":   "1.3G",
	},
}

func TestEval(t *testing.T) {
	tests := []struct {
		name    string
		exp     string
		want    any
		wantErr bool
		skip    bool
	}{
		{
			name: "lte",
			exp:  "object.replicas <= 5",
			want: true,
		},
		{
			name:    "error",
			exp:     "object.",
			wantErr: true,
		},
		{
			name: "url",
			exp:  "isURL(object.href) && url(object.href).getScheme() == 'https' && url(object.href).getEscapedPath() == '/path'",
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/2
		},
		{
			name: "query",
			exp:  "url(object.href).getQuery()",
			want: map[string]any{
				"query": []any{"val"},
			},
			skip: true, // https://github.com/polds/expr-playground/issues/3
		},
		{
			name: "regex",
			exp:  "object.image.find('v[0-9]+.[0-9]+.[0-9]*$')",
			want: "v0.0.0",
			skip: true, // https://github.com/polds/expr-playground/issues/4
		},
		{
			name: "list",
			// For some reason object.items == sort(object.items) is false. Needs further investigation.
			exp:  `object.items == sort(object.items) && sum(object.items) == 6 && sort(object.items)[-1] == 3 && findIndex(object.items, # == 1) == 0`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/5
		},
		{
			name: "optional",
			exp:  `object?.foo ?? "fallback"`,
			want: "fallback",
		},
		{
			name: "strings",
			exp:  "join(object.abc, ', ')",
			want: "a, b, c",
		},
		{
			name: "cross type numeric comparisons",
			exp:  "object.replicas > 1.4",
			want: true,
		},
		{
			name: "split pipe",
			exp:  "object.image | split(':') | len() == 2",
			want: true,
		},
		{
			name: "split",
			exp:  "object.image | split() | len() == 2",
			want: true,
		},
		{
			name: "quantity",
			exp:  `isQuantity(object.memory) && quantity(object.memory).add(quantity("700M")).sub(1).isLessThan(quantity("2G"))`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/9
		},
		{
			name: "duration",
			exp:  `duration('1d') != duration('1d')`,
			want: true,
		},
		{
			name: "sets.contains test 1",
			exp:  `sets.contains([], [])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/10
		},
		{
			name: "sets.contains test 2",
			exp:  `sets.contains([], [1])`,
			want: false,
			skip: true, // https://github.com/polds/expr-playground/issues/11
		},
		{
			name: "sets.contains test 3",
			exp:  `sets.contains([1, 2, 3, 4], [2, 3])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/12
		},
		{
			name: "sets.contains test 4",
			exp:  `sets.contains([1, 2, 3], [3, 2, 1])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/13
		},
		{
			name: "sets.equivalent test 1",
			exp:  `sets.equivalent([], [])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/14
		},
		{
			name: "sets.equivalent test 2",
			exp:  `sets.equivalent([1], [1, 1])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/15
		},
		{
			name: "sets.equivalent test 3",
			exp:  `sets.equivalent([1], [1, 1])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/16
		},
		{
			name: "sets.equivalent test 4",
			exp:  `sets.equivalent([1, 2, 3], [3, 2, 1])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/17
		},

		{
			name: "sets.intersects test 1",
			exp:  `sets.intersects([1], [])`,
			want: false,
			skip: true, // https://github.com/polds/expr-playground/issues/18
		},
		{
			name: "sets.intersects test 2",
			exp:  `sets.intersects([1], [1, 2])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/19
		},
		{
			name: "sets.intersects test 3",
			exp:  `sets.intersects([[1], [2, 3]], [[1, 2], [2, 3]])`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/20
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping broken test due to CEL -> Expr migration.")
			}

			got, err := Eval(tt.exp, input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() got error = %v, wantErr %t", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var res RunResponse
			if err := json.Unmarshal([]byte(got), &res); err != nil {
				t.Fatalf("json.Unmarshal got error = %v, want %v", err, nil)
			}
			if diff := cmp.Diff(tt.want, res.Result); diff != "" {
				t.Errorf("Eval() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestValidation compiles the expr expression and then runs it against the given input.
func TestValidation(t *testing.T) {
	tests := []struct {
		name          string
		exp           string
		shouldCompile bool
		shouldRun     bool
		skip          bool
	}{
		{
			name: "garbage",
			exp:  `blue ~!< 1`,
		},
		// Duration Literals
		{
			name:          "Duration Validation test 1",
			exp:           `duration('1')`, // Missing unit in duration.
			shouldCompile: true,
		},
		{
			// Note: This duration is invalid. There are two things going on here:
			//  - In vanilla go this is a rune literal and is invalid because it's more than one character.
			//  - But the Expr compilation type casts the any to a string and go parses it as a string. But this is
			// 	  still invalid because day is not a duration valid unit. So Expr compiles is successfully, but
			//    will fail Eval. So we run the program after compilation to catch this.
			//
			// https://github.com/expr-lang/expr/blob/master/builtin/builtin.go#L615
			name:          "Duration Validation test 2",
			exp:           `duration('1d')`, // No day unit in Go duration.
			shouldCompile: true,
		},
		{
			name:          "Duration Validation test 3",
			exp:           `duration('1us') < duration('1nns')`,
			shouldCompile: true,
		},
		{
			name:          "Duration Validation test 4",
			exp:           `duration('2h3m4s5us')`,
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Duration Validation test 5",
			exp:           `duration(x)`,
			shouldCompile: true,
		},

		// Timestamp Literals
		{
			name:          "Timestamp Validation test 1",
			exp:           `date('1000-00-00T00:00:00Z')`, // This is an invalid date.
			shouldCompile: true,
		},
		{
			name:          "Timestamp Validation test 2",
			exp:           `date('1000-01-01T00:00:00Z')`, // This is a valid date. But it's the min date.
			shouldCompile: true,
			shouldRun:     true,
		},

		{
			name:          "Timestamp Validation test 3",
			exp:           `date('1000-01-01T00:00:00ZZ')`, // Two Z's is invalid.
			shouldCompile: true,
		},
		{
			name: "Timestamp Validation test 4",
			exp:  `date(-6213559680)`, // unit missing
		},
		{
			name:          "Timestamp Validation test 5",
			exp:           `date(x)`,
			shouldCompile: true,
		},

		// Regex Literals
		{
			name: "Regex Validation test 1",
			exp:  `'hello'.matches('el*')`,
		},
		{
			name: "Regex Validation test 2",
			exp:  `'hello'.matches('x++')`,
		},
		{
			name: "Regex Validation test 3",
			exp:  `'hello'.matches('(?<name%>el*)')`,
		},
		{
			name: "Regex Validation test 4",
			exp:  `'hello'.matches('??el*')`,
		},
		{
			name: "Regex Validation test 5",
			exp:  `'hello'.matches(x)`,
		},

		// Homogeneous Aggregate Literals
		{
			name:          "Homogeneous Aggregate Validation test 1",
			exp:           `name in ['hello', 0]`, // Expr allows the type mixed array.
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 2",
			exp:           `{'hello':'world', 1:'!'}`, // Expr casts the integer to a string literal. This may be a bug.
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 3",
			exp:           `name in {'hello':'world', 'goodbye':true}`, // Expr correctly handles the boolean value.
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 4",
			exp:           `name in ['hello', 'world']`, // This is a valid Expr expression.
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 5",
			exp:           `name in ['hello', optional ?? '']`,
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 7",
			exp:           `name in {'hello': false, 'world': true}`, // Expr allows this even though the compared type is a string.
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 8",
			exp:           `{'hello': false, 'world': optional ?? true }`,
			shouldCompile: true,
			shouldRun:     true,
		},
		{
			name:          "Homogeneous Aggregate Validation test 9",
			exp:           `{'hello': optional ?? false, 'world': true}`,
			shouldCompile: true,
			shouldRun:     true,
		},
	}

	env := map[string]any{
		"x":    "",
		"name": "",
	}
	opts := append(exprEnvOptions, expr.Env(env), expr.AllowUndefinedVariables())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping broken test due to CEL -> Expr migration.")
			}

			program, err := expr.Compile(tt.exp, opts...)
			if (err != nil) == tt.shouldCompile {
				t.Errorf("Compile() got error = %v, shouldCompile %t", err, tt.shouldCompile)
			}

			_, err = expr.Run(program, env)
			if (err != nil) == tt.shouldRun {
				t.Errorf("Run() got error = %v, shouldRun %t", err, tt.shouldRun)
			}
		})
	}
}
