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
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/expr-lang/expr"
)

// IsWeb3Checksummed is a function that checks whether the given address (or list of addresses) is checksummed. It is provided as an Expr function.
// It supports the following types:
// - string
// - []any (which should contain only string elements)

// Examples:
// - isWeb3Checksummed("0xb0F001C7F6C665b7b8e12F29EDC1107613fe980D")
// - isWeb3Checksummed(["0xb0F001C7F6C665b7b8e12F29EDC1107613fe980D", "0x3106E2e148525b3DB36795b04691D444c24972fB"])
func IsWeb3Checksummed() expr.Option {
	return expr.Function("isWeb3Checksummed", func(params ...any) (any, error) {
		return isWeb3Checksummed(params[0])
	},
		new(func([]any) (bool, error)),
		new(func(string) (bool, error)),
	)
}

func isWeb3Checksummed(v any) (any, error) {
	if v == nil {
		return false, nil
	}

	switch t := v.(type) {
	case []any:
		return arrayChecksummed(t)
	case string:
		return checksummed(t)
	default:
		return false, fmt.Errorf("type %s is not supported", reflect.TypeOf(v))
	}
}

func arrayChecksummed(v []any) (bool, error) {
	switch t := v[0].(type) {
	case string:
		for _, address := range v {
			res, err := checksummed(address.(string))
			if err != nil || !res {
				return res, err
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported type %T", t)
	}
}

func checksummed(address string) (bool, error) {
	if len(address) != 42 {
		return false, fmt.Errorf("address needs to be 42 characters long")
	}

	if !strings.HasPrefix(address, "0x") {
		return false, fmt.Errorf("address needs to start with 0x")
	}

	return common.IsHexAddress(address) && checksumAddress(address) == address, nil
}

// Algorithm for checksumming a web3 address:
// - Convert the address to lowercase
// - Hash the address using keccak256
// - Take 40 characters of the hash, drop the rest (40 because of the address length)
// - Iterate through each character in the original address
//   - If the checksum character >= 8 and character in the original address at the same idx is [a, f] then capitalize
//   - Otherwise, add character
//
// For visualization, you can watch the following video: https://www.youtube.com/watch?v=2vH_CQ_rvbc
func checksumAddress(address string) string {
	if strings.HasPrefix(address, "0x") {
		address = address[2:]
	}

	lowercaseAddress := strings.ToLower(address)
	hashedAddress := crypto.Keccak256([]byte(lowercaseAddress))
	checksum := hex.EncodeToString(hashedAddress)[:40]

	var checksumAddress strings.Builder
	for idx, char := range lowercaseAddress {
		if checksum[idx] >= '8' && (char >= 'a' && char <= 'f') {
			checksumAddress.WriteRune(char - 32)
		} else {
			checksumAddress.WriteRune(char)
		}
	}

	return "0x" + checksumAddress.String()
}
