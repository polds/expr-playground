// Copyright 2024 Peter Olds <me@polds.dev>
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

package functions

import (
	"fmt"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s: %d", p.Name, p.Age)
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

func Test_isSorted(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		sorted, err := isSorted(nil)
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("sort.Interface - not sorted", func(t *testing.T) {
		people := []Person{
			{"Bob", 31},
			{"John", 42},
			{"Michael", 17},
			{"Jenny", 26},
		}
		sorted, err := isSorted(ByAge(people))
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("sort.Interface - sorted", func(t *testing.T) {
		people := []Person{
			{"Michael", 17},
			{"Jenny", 26},
			{"Bob", 31},
			{"John", 42},
		}
		sorted, err := isSorted(ByAge(people))
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("any - int slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{5, 4, 3, 2, 1})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - int slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{1, 2, 3, 4, 5})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("any - mis-typed int slice", func(t *testing.T) {
		sorted, err := isSorted([]any{1, 2, 3, "4", 5})
		require.Error(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - float slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{5.0, 4.0, 3.0, 2.0, 1.0})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - float slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{1.0, 2.0, 3.0, 4.0, 5.0})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("any - mis-typed float slice", func(t *testing.T) {
		sorted, err := isSorted([]any{1.0, 2.0, 3.0, "4.0", 5.0})
		require.Error(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - string slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{"5", "4", "3", "2", "1"})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - string slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]any{"1", "2", "3", "4", "5"})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("any - mis-typed string slice", func(t *testing.T) {
		sorted, err := isSorted([]any{"1", "2", "3", 4, "5"})
		require.Error(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("any - unsupported type", func(t *testing.T) {
		sorted, err := isSorted([]any{Person{"Bob", 31}})
		require.Error(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("int slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]int{5, 4, 3, 2, 1})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("int slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]int{1, 2, 3, 4, 5})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("float slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]float64{5.0, 4.0, 3.0, 2.0, 1.0})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("float slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("string slice - not sorted", func(t *testing.T) {
		sorted, err := isSorted([]string{"5", "4", "3", "2", "1"})
		require.NoError(t, err)
		assert.False(t, sorted.(bool))
	})
	t.Run("string slice - sorted", func(t *testing.T) {
		sorted, err := isSorted([]string{"1", "2", "3", "4", "5"})
		require.NoError(t, err)
		assert.True(t, sorted.(bool))
	})
	t.Run("unsupported type", func(t *testing.T) {
		sorted, err := isSorted(Person{"Bob", 31})
		require.Error(t, err)
		assert.False(t, sorted.(bool))
	})
}

func TestIsSorted(t *testing.T) {
	tests := []struct {
		name           string
		exp            string
		want           bool
		wantCompileErr bool
		wantRuntimeErr bool
	}{
		{
			name: "nil",
			exp:  `isSorted(nil)`,
			want: false,
		},
		{
			name: "sort.Interface - not sorted",
			exp:  `isSorted(people_unsorted)`,
		},
		{
			name: "sort.Interface - sorted",
			exp:  `isSorted(people_sorted)`,
			want: true,
		},
		{
			name: "int slice - not sorted",
			exp:  `isSorted(ints_unsorted)`,
		},
		{
			name: "int slice - sorted",
			exp:  `isSorted(ints_sorted)`,
			want: true,
		},
		{
			name: "float slice - not sorted",
			exp:  `isSorted(floats_unsorted)`,
		},
		{
			name: "float slice - sorted",
			exp:  `isSorted(floats_sorted)`,
			want: true,
		},
		{
			name: "string slice - not sorted",
			exp:  `isSorted(strings_unsorted)`,
		},
		{
			name: "string slice - sorted",
			exp:  `isSorted(strings_sorted)`,
			want: true,
		},
		{
			name: "any - int slice - not sorted",
			exp:  `isSorted(any_unsorted)`,
		},
		{
			name: "any - int slice - sorted",
			exp:  `isSorted(any_sorted)`,
			want: true,
		},
		{
			name:           "any - mis-typed int slice",
			exp:            `isSorted(any_mixed_slice)`,
			wantRuntimeErr: true,
		},
		{
			name:           "unsupported type",
			exp:            `isSorted(v)`,
			wantCompileErr: true,
		},
		{
			name:           "no argument",
			exp:            `isSorted()`,
			wantCompileErr: true,
		},
		{
			name:           "too many arguments",
			exp:            `isSorted(ints_sorted, ints_sorted)`,
			wantCompileErr: true,
		},
	}

	people := []Person{
		{"Michael", 17},
		{"Jenny", 26},
		{"Bob", 31},
		{"John", 42},
	}
	input := map[string]any{
		"people_sorted": ByAge(people),
		"people_unsorted": func() ByAge {
			ii := make([]Person, len(people))
			copy(ii, people)
			ii[0], ii[1] = ii[1], ii[0]
			return ii
		}(),
		"ints_sorted":      []int{1, 2, 3, 4, 5},
		"ints_unsorted":    []int{5, 4, 3, 2, 1},
		"floats_sorted":    []float64{1.0, 2.0, 3.0, 4.0, 5.0},
		"floats_unsorted":  []float64{5.0, 4.0, 3.0, 2.0, 1.0},
		"strings_sorted":   []string{"1", "2", "3", "4", "5"},
		"strings_unsorted": []string{"5", "4", "3", "2", "1"},
		"any_unsorted":     []any{5, 4, 3, 2, 1},
		"any_sorted":       []any{1, 2, 3, 4, 5},
		"any_mixed_slice":  []any{1, 2, 3, "4", 5},
		"v":                true,
	}
	opts := []expr.Option{
		expr.Env(input),
		expr.AsBool(),
		expr.DisableAllBuiltins(),
		IsSorted(),
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			program, err := expr.Compile(tc.exp, opts...)
			if tc.wantCompileErr && err == nil {
				require.Error(t, err)
			}
			if !tc.wantCompileErr && err != nil {
				require.NoError(t, err)
			}
			if tc.wantCompileErr {
				return
			}

			got, err := expr.Run(program, input)
			if tc.wantRuntimeErr && err == nil {
				require.Error(t, err)
			}
			if !tc.wantRuntimeErr && err != nil {
				require.NoError(t, err)
			}
			if tc.wantRuntimeErr {
				return
			}
			assert.IsType(t, tc.want, got)
			assert.Equal(t, tc.want, got)
		})
	}
}
