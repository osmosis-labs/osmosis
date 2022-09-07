package simulation

import (
	"errors"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v12/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v12/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomMsgCreateDenom creates a random tokenfactory denom that is no greater than 44 alphanumeric characters
func RandomMsgCreateDenom(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgCreateDenom, error) {
	minCoins := k.GetParams(ctx).DenomCreationFee
	acc, err := sim.RandomSimAccountWithMinCoins(ctx, minCoins)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDenom{
		Sender:   acc.Address.String(),
		Subdenom: sim.RandStringOfLength(types.MaxSubdenomLength),
	}, nil
}

// RandomMsgMintDenom takes a random denom that has been created and uses the denom's admin to mint a random amount
func RandomMsgMintDenom(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgMint, error) {
	acc, senderExists := sim.RandomSimAccountWithConstraint(accountCreatedTokenFactoryDenom(k, ctx))
	if !senderExists {
		return nil, errors.New("no addr has created a tokenfactory coin")
	}

	denom, addr, err := getTokenFactoryDenomAndItsAdmin(k, sim, ctx, acc)
	if err != nil {
		return nil, err
	}
	if addr == nil {
		return nil, errors.New("denom has no admin")
	}

	// TODO: Replace with an improved rand exponential coin
	mintAmount := sim.RandPositiveInt(sdk.NewIntFromUint64(1000_000000))
	return &types.MsgMint{
		Sender: addr.String(),
		Amount: sdk.NewCoin(denom, mintAmount),
	}, nil
}

// RandomMsgBurnDenom takes a random denom that has been created and uses the denom's admin to burn a random amount
func RandomMsgBurnDenom(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgBurn, error) {
	acc, senderExists := sim.RandomSimAccountWithConstraint(accountCreatedTokenFactoryDenom(k, ctx))
	if !senderExists {
		return nil, errors.New("no addr has created a tokenfactory coin")
	}

	denom, addr, err := getTokenFactoryDenomAndItsAdmin(k, sim, ctx, acc)
	if err != nil {
		return nil, err
	}
	if addr == nil {
		return nil, errors.New("denom has no admin")
	}

	denomBal := sim.BankKeeper().GetBalance(ctx, addr, denom)
	if denomBal.IsZero() {
		return nil, errors.New("addr does not have enough balance to burn")
	}

	// TODO: Replace with an improved rand exponential coin
	burnAmount := sim.RandPositiveInt(denomBal.Amount)
	return &types.MsgBurn{
		Sender: addr.String(),
		Amount: sdk.NewCoin(denom, burnAmount),
	}, nil
}

// RandomMsgChangeAdmin takes a random denom that has been created and changes the admin to another random account
func RandomMsgChangeAdmin(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgChangeAdmin, error) {
	acc, senderExists := sim.RandomSimAccountWithConstraint(accountCreatedTokenFactoryDenom(k, ctx))
	if !senderExists {
		return nil, errors.New("no addr has created a tokenfactory coin")
	}

	denom, addr, err := getTokenFactoryDenomAndItsAdmin(k, sim, ctx, acc)
	if err != nil {
		return nil, err
	}
	if addr == nil {
		return nil, errors.New("denom has no admin")
	}

	newAdmin := sim.RandomSimAccount()
	if newAdmin.Address.String() == addr.String() {
		return nil, errors.New("new admin cannot be the same as current admin")
	}

	return &types.MsgChangeAdmin{
		Sender:   addr.String(),
		Denom:    denom,
		NewAdmin: newAdmin.Address.String(),
	}, nil
}

func accountCreatedTokenFactoryDenom(k keeper.Keeper, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		store := k.GetCreatorPrefixStore(ctx, acc.Address.String())
		iterator := store.Iterator(nil, nil)
		defer iterator.Close()
		return iterator.Valid()
	}
}

func getTokenFactoryDenomAndItsAdmin(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context, acc legacysimulationtype.Account) (string, sdk.AccAddress, error) {
	store := k.GetCreatorPrefixStore(ctx, acc.Address.String())
	denoms := osmoutils.GatherAllKeysFromStore(store)
	denom := simtypes.RandSelect(sim, denoms...)

	authData, err := k.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return "", nil, err
	}
	admin := authData.Admin
	addr, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return "", nil, err
	}
	return denom, addr, nil
}
