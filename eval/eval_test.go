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
			exp:  "object.items.isSorted() && object.items.sum() == 6 && object.items.max() == 3 && object.items.indexOf(1) == 0",
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/5
		},
		{
			name: "optional",
			exp:  `object.?foo.orValue("fallback")`,
			want: "fallback",
			skip: true, // https://github.com/polds/expr-playground/issues/6
		},
		{
			name: "strings",
			exp:  "object.abc.join(', ')",
			want: "a, b, c",
			skip: true, // https://github.com/polds/expr-playground/issues/7
		},
		{
			name: "cross type numeric comparisons",
			exp:  "object.replicas > 1.4",
			want: true,
		},
		{
			name: "split",
			exp:  "object.image.split(':').size() == 2",
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/8
		},
		{
			name: "quantity",
			exp:  `isQuantity(object.memory) && quantity(object.memory).add(quantity("700M")).sub(1).isLessThan(quantity("2G"))`,
			want: true,
			skip: true, // https://github.com/polds/expr-playground/issues/9
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

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		exp     string
		wantErr bool
		skip    bool
	}{
		// Duration Literals
		{
			name:    "Duration Validation test 1",
			exp:     `duration('1')`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/21
		},
		{
			name:    "Duration Validation test 2",
			exp:     `duration('1d')`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/22
		},
		{
			name:    "Duration Validation test 3",
			exp:     `duration('1us') < duration('1nns')`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/23
		},
		{
			name: "Duration Validation test 4",
			exp:  `duration('2h3m4s5us')`,
		},
		{
			name: "Duration Validation test 5",
			exp:  `duration(x)`,
		},

		// Timestamp Literals
		{
			name:    "Timestamp Validation test 1",
			exp:     `timestamp('1000-00-00T00:00:00Z')`,
			wantErr: true,
		},
		{
			name:    "Timestamp Validation test 2",
			exp:     `timestamp('1000-01-01T00:00:00ZZ')`,
			wantErr: true,
		},
		{
			name: "Timestamp Validation test 3",
			exp:  `timestamp('1000-01-01T00:00:00Z')`,
			skip: true, // https://github.com/polds/expr-playground/issues/24
		},
		{
			name: "Timestamp Validation test 4",
			exp:  `timestamp(-6213559680)`, // min unix epoch time.
			skip: true,                     // https://github.com/polds/expr-playground/issues/25
		},
		{
			name:    "Timestamp Validation test 5",
			exp:     `timestamp(-62135596801)`,
			wantErr: true,
		},
		{
			name: "Timestamp Validation test 6",
			exp:  `timestamp(x)`,
			skip: true, // https://github.com/polds/expr-playground/issues/26
		},

		// Regex Literals
		{
			name: "Regex Validation test 1",
			exp:  `'hello'.matches('el*')`,
			skip: true, // https://github.com/polds/expr-playground/issues/27
		},
		{
			name:    "Regex Validation test 2",
			exp:     `'hello'.matches('x++')`,
			wantErr: true,
		},
		{
			name:    "Regex Validation test 3",
			exp:     `'hello'.matches('(?<name%>el*)')`,
			wantErr: true,
		},
		{
			name:    "Regex Validation test 4",
			exp:     `'hello'.matches('??el*')`,
			wantErr: true,
		},
		{
			name: "Regex Validation test 5",
			exp:  `'hello'.matches(x)`,
			skip: true, // https://github.com/polds/expr-playground/issues/28
		},

		// Homogeneous Aggregate Literals
		{
			name:    "Homogeneous Aggregate Validation test 1",
			exp:     `name in ['hello', 0]`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/29
		},
		{
			name:    "Homogeneous Aggregate Validation test 2",
			exp:     `{'hello':'world', 1:'!'}`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/30
		},
		{
			name:    "Homogeneous Aggregate Validation test 3",
			exp:     `name in {'hello':'world', 'goodbye':true}`,
			wantErr: true,
			skip:    true, // https://github.com/polds/expr-playground/issues/31
		},
		{
			name: "Homogeneous Aggregate Validation test 4",
			exp:  `name in ['hello', 'world']`,
			skip: true, // https://github.com/polds/expr-playground/issues/31
		},
		{
			name: "Homogeneous Aggregate Validation test 5",
			exp:  `name in ['hello', ?optional.ofNonZeroValue('')]`,
			skip: true, // https://github.com/polds/expr-playground/issues/32
		},
		{
			name: "Homogeneous Aggregate Validation test 6",
			exp:  `name in [?optional.ofNonZeroValue(''), 'hello', ?optional.of('')]`,
			skip: true, // https://github.com/polds/expr-playground/issues/33
		},
		{
			name: "Homogeneous Aggregate Validation test 7",
			exp:  `name in {'hello': false, 'world': true}`,
		},
		{
			name: "Homogeneous Aggregate Validation test 8",
			exp:  `{'hello': false, ?'world': optional.ofNonZeroValue(true)}`,
			skip: true, // https://github.com/polds/expr-playground/issues/34
		},
		{
			name: "Homogeneous Aggregate Validation test 9",
			exp:  `{?'hello': optional.ofNonZeroValue(false), 'world': true}`,
			skip: true, // https://github.com/polds/expr-playground/issues/35
		},
	}

	env := map[string]any{
		"x":    "",
		"name": "",
	}
	opts := append(exprEnvOptions, expr.Env(env))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping broken test due to CEL -> Expr migration.")
			}

			_, err := expr.Compile(tt.exp, opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() got error = %v, wantErr %t", err, tt.wantErr)
				return
			}
		})
	}
}
