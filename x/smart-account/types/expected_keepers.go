package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ContractKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
