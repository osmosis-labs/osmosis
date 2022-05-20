package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) CreateDenom(ctx sdk.Context, creatorAddr string, denomNonce string) (newTokenDenom string, err error) {
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

	denom, err := types.GetTokenDenom(creatorAddr, denomNonce)
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
	err = k.SetAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return "", err
	}

	k.addDenomFromCreator(ctx, creatorAddr, denom)
	return denom, nil
}
