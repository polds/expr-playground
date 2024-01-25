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
	"testing"

	"github.com/expr-lang/expr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWeb3Checksummed(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		want           bool
		wantCompileErr bool
		wantRuntimeErr bool
	}{
		{
			name: "nil",
			expr: `isWeb3Checksummed(nil)`,
			want: false,
		},
		{
			name: "string - not checksummed",
			expr: `isWeb3Checksummed('0x30F4283a3d6302f968909Ff7c02ceCB2ac6C27Ac')`,
			want: false,
		},
		{
			name: "string - checksummed",
			expr: `isWeb3Checksummed('0x30D873664Ba766C983984C7AF9A921ccE36D34e1')`,
			want: true,
		},
		{
			name: "string slice - checksummed",
			expr: `isWeb3Checksummed(['0x55028780918330FD00a34a61D9a7Efd3f43ca845', '0xAA95A3e367b427477bAdAB3d104f7D04ba158895'])`,
			want: true,
		},
		{
			name: "string slice - checksummed",
			expr: `isWeb3Checksummed(['0x869C8ADA0fb9AfC753159b7D6D72Cc8bf58e6987', '0x2a92BCecd6e702702864E134821FD2DE73C3e180'])`,
			want: false,
		},
		{
			name:           "address needs to start with 0x",
			expr:           `isWeb3Checksummed('0034B03Cb9086d7D758AC55af71584F81A598759FE')`,
			wantRuntimeErr: true,
		},
		{
			name:           "address needs to be 42 characters long",
			expr:           `isWeb3Checksummed('34B03Cb9086d7D758AC55af71584F81A598759FE')`,
			wantRuntimeErr: true,
		},
		{
			name:           "unsupported type int",
			expr:           `isWeb3Checksummed(0)`,
			wantCompileErr: true,
		},
		{
			name:           "unsupported type int",
			expr:           `isWeb3Checksummed([0])`,
			wantRuntimeErr: true,
		},
		{
			name:           "not enough arguments",
			expr:           `isWeb3Checksummed()`,
			wantCompileErr: true,
		},
	}

	opts := []expr.Option{
		expr.AsBool(),
		expr.DisableAllBuiltins(),
		IsWeb3Checksummed(),
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			program, err := expr.Compile(tc.expr, opts...)
			if tc.wantCompileErr && err == nil {
				require.Error(t, err)
			}
			if !tc.wantCompileErr && err != nil {
				require.NoError(t, err)
			}
			if tc.wantCompileErr {
				return
			}

			got, err := expr.Run(program, nil)
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
