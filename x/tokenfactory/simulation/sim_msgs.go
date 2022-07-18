package simulation

import (
	"errors"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomMsgCreateDenom creates a random tokenfactory denom that is no greater than 44 alphanumeric characters
func RandomMsgCreateDenom(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgCreateDenom, error) {
	// select a pseudo-random simulation account
	sender := sim.RandomSimAccount()

	subdenom := sim.RandStringOfLength(44)

	return &types.MsgCreateDenom{
		Sender:   sender.Address.String(),
		Subdenom: subdenom,
	}, nil
}

func RandomMsgMintDenom(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgMint, error) {
	var tokenFactoryMint sdk.Coin
	acc, senderExists := sim.RandomSimAccountWithConstraint(accountHasTokenFactoryDenomConstraint(k, ctx))
	if !senderExists {
		return nil, errors.New("no addr has created a tokenfactory coin")
	}
	store := k.GetCreatorPrefixStore(ctx, acc.Address.String())

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	denoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		denoms = append(denoms, string(iterator.Key()))
	}
	randIndex := simulation.RandLTBound(sim, len(denoms))
	denom := denoms[randIndex]
	tokenFactoryMint.Denom = denom
	tokenFactoryMint.Amount = sdk.NewInt(int64(sim.RandIntBetween(0, 1000000000000)))
	return &types.MsgMint{
		Sender: acc.Address.String(),
		Amount: tokenFactoryMint,
	}, nil
}

func accountHasTokenFactoryDenomConstraint(k keeper.Keeper, ctx sdk.Context) simulation.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		store := k.GetCreatorPrefixStore(ctx, acc.Address.String())

		iterator := store.Iterator(nil, nil)
		defer iterator.Close()

		denoms := []string{}
		for ; iterator.Valid(); iterator.Next() {
			denoms = append(denoms, string(iterator.Key()))
		}
		return len(denoms) != 0
	}
}
