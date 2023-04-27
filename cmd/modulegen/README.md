# Module gen

This package contains code for generating osmosis proto and module logic.

## Running it

This should be run in the osmosis root directory, as follows:

```bash
make run-modulegen
```

or

```bash
go run cmd/modulegen/main.go -module_name test_module
make proto-gen
```
