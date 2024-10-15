package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the module name.
	ModuleName = "callback"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

var (
	ParamsKeyPrefix   = collections.NewPrefix(1)
	CallbackKeyPrefix = collections.NewPrefix(2)
)
