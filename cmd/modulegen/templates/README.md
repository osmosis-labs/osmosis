# Querygen

This package contains code for generating osmosis SDK module boilerplate.

It does not do any wiring. As a result, this ends up being non-state-breaking.

It recursively searches the proto directory for `query.yml` files, and then builds generated grpc, cli and proto wrapping code.

## Running it

This should be run in the osmosis root directory, as either:

```bash
make run-querygen
```

or

```bash
go run cmd/querygen/main.go
```
