package keeper

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type GammHooks struct {
	k Keeper
}

type EpochHooks struct {
	k Keeper
}

var (
	_ gammtypes.GammHooks    = GammHooks{}
	_ epochstypes.EpochHooks = EpochHooks{}
)

func (k Keeper) EpochHooks() epochstypes.EpochHooks {
	return EpochHooks{k}
}

// Create new pool incentives hooks.
func (k Keeper) GAMMHooks() gammtypes.GammHooks {
	return GammHooks{k}
}

// AfterPoolCreated creates a gauge for each pool’s lockable duration.
func (h GammHooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	fmt.Println("AfterPoolCreated: IN THE HOOK")

	// POOL

	fmt.Println("\n_______________POOL TINGS_______________")

	pool, err := h.k.gammKeeper.GetPoolAndPoke(ctx, poolId)

	if err != nil {
		fmt.Println("\nAfterPoolCreated: ERROR GETTING POOL")
	}

	fmt.Println("Created Pool Liquidity:", pool.GetTotalPoolLiquidity(ctx))

	// Module Address / Account Setup

	fmt.Println("\n_______________MODULE ACCOUNT TINGS_______________")

	protorevModuleAddress, permissions := h.k.accountKeeper.GetModuleAddressAndPermissions(types.ModuleName)

	fmt.Println("\nprotorevModuleAddress:", protorevModuleAddress)
	fmt.Println("permissions:", permissions)

	protorevModuleAccount := h.k.accountKeeper.GetAccount(ctx, protorevModuleAddress)

	if protorevModuleAccount == nil {
		moduleAcc := authtypes.NewEmptyModuleAccount(
			types.ModuleName, authtypes.Minter, authtypes.Burner)

		h.k.accountKeeper.SetAccount(ctx, moduleAcc)

		protorevModuleAccount = h.k.accountKeeper.GetAccount(ctx, protorevModuleAddress)
	}

	fmt.Println("protorevModuleAccount:", protorevModuleAccount)

	// Calc Tokens in

	tokenInCoins := pool.GetTotalPoolLiquidity(ctx)[0]
	tokenInCoins.Amount = tokenInCoins.Amount.Quo(sdk.NewInt(10))

	tokensIn := sdk.NewCoins(tokenInCoins)

	// Mint Coins

	fmt.Println("\n_______________MINT COINS TINGS_______________")

	balancesBefore := h.k.bankKeeper.GetAllBalances(ctx, protorevModuleAddress)

	fmt.Println("\nBefore Mint Balance:", balancesBefore)

	mintCoinError := h.k.bankKeeper.MintCoins(ctx, types.ModuleName, tokensIn)

	if mintCoinError != nil {
		fmt.Println("Error minting coins:", mintCoinError)
	}

	balancesAfter := h.k.bankKeeper.GetAllBalances(ctx, protorevModuleAddress)

	fmt.Println("After Mint Balance:", balancesAfter)

	// SWAP

	fmt.Println("\n_______________SWAP TINGS_______________")

	tokenOutDenom := pool.GetTotalPoolLiquidity(ctx)[1].Denom

	fmt.Println("\nTOKEN IN FOR SWAP:", tokensIn)
	fmt.Println("TOKEN OUT DENOM FOR SWAP:", tokenOutDenom)

	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, pool.GetSwapFee(ctx))

	fmt.Println("TOKEN OUT SHOULD BE:", tokenOut)

	tokenIn := tokensIn[0]

	tokenOutAmount, swapErr := h.k.gammKeeper.SwapExactAmountIn(ctx, protorevModuleAddress, poolId, tokenIn, tokenOutDenom, sdk.NewInt(1))

	fmt.Println(swapErr)
	fmt.Println(tokenOutAmount)

	if swapErr != nil {
		fmt.Println("Error swapping coins:", swapErr)
	}

	fmt.Println("TokenOutAmount on Real Swap:", tokenOutAmount)

	balancesAfterSwap := h.k.bankKeeper.GetAllBalances(ctx, protorevModuleAddress)

	fmt.Println("Balance After Swap:", balancesAfterSwap)

	poolAfterSwap, errAfterSwap := h.k.gammKeeper.GetPoolAndPoke(ctx, poolId)

	if errAfterSwap != nil {
		fmt.Println("Error getting pool after swap:", errAfterSwap)
	}

	fmt.Println("Pool After Swap:", poolAfterSwap.GetTotalPoolLiquidity(ctx))

	// GAS

	fmt.Println("\n_______________GAS TINGS_______________")

	fmt.Println("\nBefore Refund:", ctx.GasMeter())

	ctx.GasMeter().RefundGas(ctx.GasMeter().GasConsumedToLimit(), "refund gas")

	fmt.Println("After Refund", ctx.GasMeter())

	// GAS TESTS

	fmt.Println("\n_______________GAS TEST TINGS_______________")

	start := time.Now()
	poolsRes, poolResErr := h.k.gammKeeper.GetPoolsAndPoke(ctx)

	if poolResErr != nil {
		fmt.Println("Error getting pools:", poolResErr)
	}
	end := time.Now()
	//fmt.Println("Pools:", poolsRes)
	//fmt.Println(len(poolsRes))
	fmt.Println("Time to get pools:", end.Sub(start))

	fmt.Println("Before transient store set:", ctx.GasMeter())

	fmt.Println("Number of bytes set:", len([]byte(fmt.Sprintf("%v", poolsRes))))

	transientStore := ctx.TransientStore(h.k.transientKey)
	transientStore.Set([]byte{1}, []byte(fmt.Sprintf("%v", poolsRes)))
	fmt.Println("After transient store set:", ctx.GasMeter())
	transientStore.Get([]byte{1})
	fmt.Println("After transient store get:", ctx.GasMeter())

	memStore := ctx.KVStore(h.k.memKey)
	memStoreType := memStore.GetStoreType()
	fmt.Println("MemStoreType:", memStoreType)
	fmt.Println("Before memory store set:", ctx.GasMeter())

	fmt.Println(memStore.Has([]byte{1}))
	memStore.Set([]byte{1}, []byte{1})
	//memStore.Set([]byte{1}, []byte(fmt.Sprintf("%v", poolsRes)))
	fmt.Println("After memory store set:", ctx.GasMeter())

	memStore.Get([]byte{1})
	fmt.Println("After memory store get:", ctx.GasMeter())

	//h.k.UpdateConnectedTokens(ctx, &allDenoms)

	//h.k.UpdateConnectedTokensToPoolIDs(ctx, allDenoms, poolId)

	//h.k.UpdatePoolRoutes(ctx, allDenoms, poolId)

	fmt.Println(ctx.EventManager().Events())
}

// AfterJoinPool hook is a noop.
func (h GammHooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
}

// AfterExitPool hook is a noop.
func (h GammHooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
}

// AfterSwap hook is a noop.
func (h GammHooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {

	needToArb := types.NeedToArb{NeedToArb: true}
	h.k.SetNeedToArb(ctx, &needToArb)

}

///////////////////////////////////////////////////////

// BeforeEpochStart is the epoch start hook.
func (hook EpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	fmt.Println("BeforeEpochStart: ", epochIdentifier, epochNumber)
	fmt.Println(hook.k.epochKeeper.GetEpochInfo(ctx, epochIdentifier))

	/*
		fmt.Println(hook.k)
		fmt.Println(hook.k.cdc)
		fmt.Println("StoreKey", hook.k.storeKey)
		fmt.Println("TransietKey", hook.k.transientKey)
		fmt.Println("MemKey:", hook.k.memKey)
		fmt.Println(hook.k.paramstore)
		fmt.Println(hook.k.accountKeeper)
		fmt.Println(hook.k.bankKeeper)
		fmt.Println(hook.k.gammKeeper)
		fmt.Println(hook.k.epochKeeper)
	*/

	return nil
}

// AfterEpochEnd is the epoch end hook.
func (hook EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	//fmt.Println("AfterEpochEnd: ", epochIdentifier, epochNumber)
	return nil
}
