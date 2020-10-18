package pool

import sdk "github.com/cosmos/cosmos-sdk/types"

type Record struct {
	DenormalizedWeight sdk.Dec
	Balance            sdk.Int
}

type Pool struct {
	Address     sdk.AccAddress
	Records     map[string]Record
	TotalWeight sdk.Int
}
