package keeper

import (
	"fmt"

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

func (k Keeper) Mint(ctx sdk.Context, denomId string, senderAddr string, amount sdk.Coin) error {
	// validate denom
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(senderAddr)
	if err != nil {
		return err
	}

	// TODO: check if the module denom is the denom that exist in the state
	denom, denomExists := k.GetDenom(ctx, denomId)
	if !denomExists {
		return fmt.Errorf("denomID %s does not exist", denom.Id)
	}

	// TODO: check if the nft id already exist
	// Get current latest token id and increment?
	tokenIdExists := k.HasNFT(ctx, denomId, '1')
	if tokenIdExists {
		return fmt.Errorf("tokenID %s already exists", '1')
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return nil
}
