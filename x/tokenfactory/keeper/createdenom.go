package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) CreateDenom(ctx sdk.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error) {
	// Temporary check until IBC bug is sorted out
	if k.bankKeeper.HasSupply(ctx, subdenom) {
		return "", fmt.Errorf("Temporary error until IBC bug is sorted out, " +
			"can't create subdenoms that are the same as a native denom.")
	}

	// Send creation fee to community pool
	creationFee := k.GetParams(ctx).DenomCreationFee
	accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
	if err != nil {
		return "", err
	}
	if len(creationFee) > 0 {
		if err := k.distrKeeper.FundCommunityPool(ctx, creationFee, accAddr); err != nil {
			return "", err
		}
	}

	denom, err := types.GetTokenDenom(creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if found {
		return "", types.ErrDenomExists
	}

	denomMetaData := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    denom,
			Exponent: 0,
		}},
		Base: denom,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)

	authorityMetadata := types.DenomAuthorityMetadata{
		Admin: creatorAddr,
	}
	err = k.setAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return "", err
	}

	k.addDenomFromCreator(ctx, creatorAddr, denom)
	return denom, nil
}
