package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, shares are minted
// and sent to the pool creator. The shares are created using a denomination in
// the form of <swap module name>/pool/{poolID}. In addition, the x/bank metadata is updated
// to reflect the newly created GAMM share denomination.
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
	panic("not implemented")
}
