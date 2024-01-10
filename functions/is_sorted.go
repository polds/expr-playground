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
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"sort"

	"github.com/expr-lang/expr"
)

// IsSorted provides the isSorted function as an Expr function. It will verify that the provided type
// is sorted ascending. It supports the following types:
// - Injected types that support the sort.Interface
// - []int
// - []float64
// - []string
//
// Usage:
//
//	// Inject into your environment.
//	_, err := expr.Compile(`foo`, expr.Env(nil), functions.ExprIsSorted())
//
// Expression:
//
//	isSorted([1, 2, 3])
//	isSorted(["a", "b", "c"])
//	isSorted([1.0, 2.0, 3.0])
//	isSorted(myCustomType) // myCustomType must implement sort.Interface
func IsSorted() expr.Option {
	return expr.Function("isSorted", func(params ...any) (any, error) {
		if len(params) != 1 {
			return false, fmt.Errorf("expected one parameter, got %d", len(params))
		}
		return isSorted(params[0])
	},
		new(func(sort.Interface) (bool, error)),
		new(func([]any) (bool, error)),
		new(func([]int) (bool, error)),
		new(func([]float64) (bool, error)),
		new(func([]string) (bool, error)),
	)
}

// isSorted attempts to determine if v is sortable, first by determine if it satisfies the sort.Interface interface,
// then by checking if it is a slice of a sortable type. If the type is a slice of type []any pass it to the
// isSliceSorted method which builds a new slice of the correct type and validates that it is sorted.
func isSorted(v any) (any, error) {
	if v == nil {
		return false, nil
	}

	switch t := v.(type) {
	case sort.Interface:
		return sort.IsSorted(t), nil

	// There are cases where Expr is passing around an []any instead of a []int, []float64, or []string.
	// This logic will attempt to do its own sorting to determine if the slice is sorted.
	case []any:
		return isSliceSorted(t)
	case []int:
		return slices.IsSorted(t), nil
	case []float64:
		return slices.IsSorted(t), nil
	case []string:
		return slices.IsSorted(t), nil
	}
	return false, fmt.Errorf("type %s is not sortable", reflect.TypeOf(v))
}

func convertTo[E cmp.Ordered](x any) (E, error) {
	var r E
	v, ok := x.(E)
	if !ok {
		return r, fmt.Errorf("mis-typed slice, expected %T, got %T", r, x)
	}
	return v, nil
}

func less[E cmp.Ordered](vv []any) (bool, error) {
	for i := len(vv) - 1; i > 0; i-- {
		l, err := convertTo[E](vv[i-1])
		if err != nil {
			return false, err
		}
		h, err := convertTo[E](vv[i])
		if err != nil {
			return false, err
		}
		if cmp.Less(h, l) {
			return false, nil
		}
	}
	return true, nil
}

// isSliceSorted attempts to determine if v is a slice of a sortable type.
// Instead of building a slice it just walks the slice and validates that it is sorted. The first unsorted element
// causes the function to return false.
// Expr only supports int, float, and string types.
func isSliceSorted(vv []any) (bool, error) {
	// We have to peek the first element to determine the type of the slice.
	switch t := vv[0].(type) {
	case int:
		return less[int](vv)
	case float64:
		return less[float64](vv)
	case string:
		return less[string](vv)
	default:
		return false, fmt.Errorf("unsupported type %T", t)
	}
}
