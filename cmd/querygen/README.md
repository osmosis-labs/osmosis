# Querygen

This package contains code for generating osmosis GRPC queries, and queryproto logic.

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