<!--
order: 4
-->

# Messages

## MsgSwap

A MsgSwap transaction denotes the Trader's intent to swap their balance of `OfferCoin` for new denomination `AskDenom`, for both Terra<>Terra and Terra<>Luna swaps.

```go
type MsgSwap struct {
	Trader    sdk.AccAddress
	OfferCoin sdk.Coin
	AskDenom  string
}
```

## MsgSwapSend
A MsgSendSwap first performs a swap of OfferCoin into AskDenom and the sends the resulting coins to ToAddress. Tax is charged normally, as if the sender were issuing a MsgSend with the resutling coins of the swap.


```go
type MsgSwapSend struct {
	FromAddress sdk.AccAddress
	ToAddress   sdk.AccAddress 
	OfferCoin   sdk.Coin
	AskDenom    string
}
```

## Functions

### ComputeSwap

```go
func (k Keeper) ComputeSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (retDecCoin sdk.DecCoin, spread sdk.Dec, err error)
```

This function detects the swap type from the offer and ask denominations and returns:

1. The amount of asked coins that should be returned for a given `offerCoin`. This is achieved by first spot-converting `offerCoin` to µSDR and then from µSDR to the desired `askDenom` with the proper exchange rate reported from by the Oracle.

2. The spread % that should be taken as a swap fee given the swap type. Terra<>Terra swaps simply have the Tobin Tax spread fee. Terra<>Luna spreads are the greater of `MinSpread` and spread from Constant Product pricing.

If the offerCoin's denomination is the same as `askDenom`, this will raise ErrRecursiveSwap.

### ApplySwapToPool

```go
func (k Keeper) ApplySwapToPool(ctx sdk.Context, offerCoin sdk.Coin, askCoin sdk.DecCoin) error
```

This function is called during the swap to update the blockchain's measure of , `TerraPoolDelta`, when the balances of the Terra and Luna liquidity pools have changed.

Terra currencies share the same liquidity pool, so `TerraPoolDelta` remains unaltered during Terra<>Terra swaps.

For Terra<>Luna swaps, the relative sizes of the pools will be different after the swap, and `delta` will be updated with the following formulas:

For Terra to Luna, `delta = delta + offerAmount`
For Luna to Terra, `delta = delta - askAmount`
