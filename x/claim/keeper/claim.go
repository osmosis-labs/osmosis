package keeper

import (
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/osmosis/x/claim/types"
)

// GetModuleAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountBalance(ctx sdk.Context) sdk.Coin {
	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	return k.bankKeeper.GetBalance(ctx, moduleAccAddr, sdk.DefaultBondDenom)
}

// SetModuleAccountBalance set balance of airdrop module
func (k Keeper) SetModuleAccountBalance(ctx sdk.Context, amount sdk.Coin) {
	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	k.bankKeeper.SetBalances(ctx, moduleAccAddr, sdk.NewCoins(amount))
}

func (k Keeper) EndAirdrop(ctx sdk.Context) error {
	err := k.fundRemainingsToCommunity(ctx)
	if err != nil {
		return err
	}
	k.clearInitialClaimables(ctx)
	return nil
}

// ClearClaimables clear claimable amounts
func (k Keeper) clearInitialClaimables(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.ClaimableStoreKey))
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
}

// SetClaimables set claimable amount from balances object
func (k Keeper) SetInitialClaimables(ctx sdk.Context, balances []banktypes.Balance) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))
	for _, bal := range balances {
		prefixStore.Set([]byte(bal.Address), []byte(bal.Coins.String()))
	}
	return nil
}

// GetClaimables get claimables for genesis export
func (k Keeper) GetInitialClaimables(ctx sdk.Context) []banktypes.Balance {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	balances := []banktypes.Balance{}
	for ; iterator.Valid(); iterator.Next() {
		coins, err := sdk.ParseCoinsNormalized(string(iterator.Value()))
		if err != nil {
			panic(err)
		}
		addrBz := iterator.Key()[len(types.ClaimableStoreKey):]
		addr := sdk.AccAddress(addrBz)
		balances = append(balances, banktypes.Balance{
			Address: addr.String(),
			Coins:   coins,
		})
	}
	return balances
}

// GetClaimable returns claimable amount for an address
func (k Keeper) GetClaimable(ctx sdk.Context, addr string) (sdk.Coins, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ClaimableStoreKey))
	if !prefixStore.Has([]byte(addr)) {
		return sdk.Coins{}, nil
	}
	bz := prefixStore.Get([]byte(addr))
	coins, err := sdk.ParseCoinsNormalized(string(bz))
	if err != nil {
		return coins, err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return coins, err
	}

	goneTime := ctx.BlockTime().Sub(params.AirdropStart)
	if goneTime <= params.DurationUntilDecay {
		// still not the time for decay
		return coins, nil
	}

	if goneTime > params.DurationUntilDecay+params.DurationOfDecay {
		// airdrop time passed
		return sdk.Coins{}, nil
	}

	claimableCoins := sdk.Coins{}
	monthlyDecayPercent := 10
	monthDuration := time.Hour * 24 * 30
	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := goneTime - params.DurationUntilDecay
	for _, coin := range coins {
		// decayPercent = decay_percent_per_month * (time_decayed) / (1 month)
		// claimable_percent = 100 - decayPercent
		// claimable_amt = claimable_coins * claimable_percent / 100
		decayPercent := monthlyDecayPercent * int(decayTime) / int(monthDuration)
		claimablePercent := int64(100) - int64(decayPercent)
		claimableAmt := big.NewInt(0).Div(coin.Amount.Mul(sdk.NewInt(claimablePercent)).BigInt(), big.NewInt(100))
		claimableCoinsWithDecay := sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(claimableAmt))
		claimableCoins = sdk.Coins{claimableCoinsWithDecay}
	}

	return claimableCoins, nil
}

// GetClaimablePercentagePerAction returns percentage per user's action when the weight of actions are same
func GetClaimablePercentagePerAction() sdk.Dec {
	numTotalActions := len(types.Action_name)
	return sdk.NewDec(1).QuoInt64(int64(numTotalActions))
}

// GetClaimablesByActivity returns the withdrawal amount from users' airdrop amount and activity made
func (k Keeper) GetWithdrawableByActivity(ctx sdk.Context, addr string) (sdk.Coins, error) {
	coins, err := k.GetClaimable(ctx, addr)
	if err != nil {
		return coins, err
	}
	percentage := GetClaimablePercentagePerAction()
	withdrawable := sdk.Coins{}
	for _, coin := range coins {
		amount := coin.Amount.ToDec().Mul(percentage).RoundInt()
		if amount.IsPositive() {
			withdrawable = withdrawable.Add(sdk.NewCoin(coin.Denom, amount))
		}
	}
	return withdrawable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoins(ctx sdk.Context, addr string) (sdk.Coins, error) {
	coins, err := k.GetWithdrawableByActivity(ctx, addr)
	if err != nil {
		return coins, err
	}
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return coins, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, coins)
	if err != nil {
		return coins, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr),
			sdk.NewAttribute(sdk.AttributeKeyAmount, coins.String()),
		),
	})
	return coins, nil
}

// FundRemainingsToCommunity fund remainings to the community when airdrop period end
func (k Keeper) fundRemainingsToCommunity(ctx sdk.Context) error {
	moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	amt := k.GetModuleAccountBalance(ctx)
	return k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), moduleAccAddr)
}
