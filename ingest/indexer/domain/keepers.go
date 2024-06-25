package domain

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Keepers struct {
	BankKeeper bankkeeper.Keeper
}
