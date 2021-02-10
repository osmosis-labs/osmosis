package keeper

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SetClaimables set claimable amount from balances object
func (k Keeper) SetClaimables(ctx sdk.Context, balances []banktypes.Balance) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))
	for _, bal := range balances {
		bz, err := bal.Coins.MarshalJSON()
		if err != nil {
			return err
		}
		prefixStore.Set([]byte(bal.Address), bz)
	}
	return nil
}

// GetClaimable returns claimable amount for an address
func (k Keeper) GetClaimable(ctx sdk.Context, addr string) (sdk.Coins, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))
	coins := sdk.Coins{}
	if !prefixStore.Has([]byte(addr)) {
		return coins, nil
	}
	bz := prefixStore.Get([]byte(addr))
	err := json.Unmarshal(bz, &coins)
	if err != nil {
		return coins, err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return coins, err
	}

	goneTime := ctx.BlockTime().Sub(params.AirdropStart)
	if goneTime < params.DurationUntilDecay {
		return coins, nil
	}

	if goneTime > params.DurationUntilDecay+params.DurationOfDecay {
		return sdk.Coins{}, nil
	}

	claimableCoins := sdk.Coins{}
	monthlyDecayPercent := 10
	monthDuration := time.Hour * 24 * 30
	decayTime := goneTime - params.DurationUntilDecay
	for _, coin := range coins {
		decayPercent := monthlyDecayPercent * int(decayTime) / int(monthDuration)
		claimablePercent := int64(100) - int64(decayPercent)
		claimableAmt := big.NewInt(0).Div(coin.Amount.Mul(sdk.NewInt(claimablePercent)).BigInt(), big.NewInt(100))
		claimableCoin := sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(claimableAmt))
		claimableCoins = claimableCoins.Add(claimableCoin)
	}

	return claimableCoins, nil
}

// Claim remove claimable amount entry and transfer it to user's account
func (k Keeper) Claim(ctx sdk.Context, addr string) (sdk.Coins, error) {
	coins, err := k.GetClaimable(ctx, addr)
	if err != nil {
		return coins, err
	}
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return coins, err
	}

	k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, coins)

	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))
	prefixStore.Delete([]byte(addr))
	return coins, nil
}
