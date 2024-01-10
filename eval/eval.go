// Copyright 2023 Undistro Authors
// Modifications Fork and conversion to Expr Copyright 2024 Peter Olds <me@polds.dev>
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
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/polds/expr-playground/functions"
)

type RunResponse struct {
	Result   any         `json:"result"`
	Bytecode []vm.Opcode `json:"bytecode"`
}

var exprEnvOptions = []expr.Option{
	expr.AsAny(),
	// Inject a custom isSorted function into the environment.
	functions.IsSorted(),
}

// Eval evaluates the expr expression against the given input.
func Eval(exp string, input map[string]any) (string, error) {
	localOpts := append([]expr.Option{expr.Env(input)}, exprEnvOptions...)
	program, err := expr.Compile(exp, localOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to compile the Expr expression: %w", err)
	}
	output, err := expr.Run(program, input)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate: %w", err)
	}

	res := &RunResponse{
		Result:   output,
		Bytecode: program.Bytecode,
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal the output: %w", err)
	}
	return string(out), nil
}
