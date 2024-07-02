package domain

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type CosmWasmPoolKeeperI interface {
}

type Keepers struct {
	BankKeeper         bankkeeper.Keeper
	CosmwasmPoolKeeper CosmWasmPoolKeeperI
}
