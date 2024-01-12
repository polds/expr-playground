# Expr Playground
![GitHub](https://img.shields.io/github/license/polds/expr-playground)
[![Go Report Card](https://goreportcard.com/badge/github.com/polds/expr-playground)](https://goreportcard.com/report/github.com/polds/expr-playground)

Expr Playground is an interactive WebAssembly (Wasm) powered environment to explore and experiment with 
[Expr-lang](https://expr-lang.org/). It provides a simple and user-friendly interface to write and quickly evaluate 
Expr expressions.

## Credits

This project is forked from [CEL Playground](https://github.com/undistro/cel-playground) and modified to support Expr. 
Please be sure to check out their project and give them a star as well!

## Expr libraries

Expr Playground is built by compiling Go code to WebAssembly. At present only the Expr engine is available in this 
environment. We will look at injecting some other utilities to make this environment more useful, on-par with the the
CEL standard library and CEL Playground.

Take a look at [all the environment options](eval/eval.go#L31).

### Playground Methods

The following custom methods are available in the playground:

#### isSorted(array)

Returns whether the list is sorted in ascending order.
```expr
isSorted([1, 2, 3]) == true
isSorted([1, 3, 2]) == false
isSorted(["apple", "banana", "cherry"]) == true
```
This custom function is importable in your own Expr code by importing github.com/polds/expr-playground/functions and
adding `functions.IsSorted()` to your environment. The library supports sorting on types that satisfy the 
`sort.Interface` interface.



## Development

Build the Wasm binary:
```shell
make build
```

Serve the static files:
```shell
make serve
```

## Contributing

We appreciate your contribution.
Please refer to our [contributing guideline](https://github.com/polds/expr-playground/blob/main/CONTRIBUTING.md) for further information.
This project adheres to the Contributor Covenant [code of conduct](https://github.com/polds/expr-playground/blob/main/CODE_OF_CONDUCT.md).

## License

Expr Playground and the original CEL Playground is available under the Apache 2.0 license. See the [LICENSE](LICENSE) file for more info.
