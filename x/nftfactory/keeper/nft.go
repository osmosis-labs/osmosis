package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) CreateDenom(ctx sdk.Context, denomId, senderAddr, denomName, denomData string) error {
	return k.SetDenom(ctx, types.Denom{
		Id:        denomId,
		Sender:    senderAddr,
		DenomName: denomName,
		Data:      denomData,
	})
}
