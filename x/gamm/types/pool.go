package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v043_temp/address"
)

type SwapI interface {
	SolveConstantFunctionInvariant(balanceA, weightA, balanceB, weightB, amountA sdk.Dec) sdk.Dec
	SolveTokenFromShare(balance, weight, totalShares, shareAmount sdk.Dec) sdk.Dec
	SolveShareFromToken(balance, weight, totalShares, tokenAmount sdk.Dec) sdk.Dec
}

// PoolI defines an interface for pools that hold tokens.
type PoolI interface {
	proto.Message

	SwapI

	GetAddress() sdk.AccAddress
	String() string

	GetId() uint64
	GetPoolSwapFee() sdk.Dec
	GetPoolExitFee() sdk.Dec
	GetTotalWeight() sdk.Int
	GetTotalShares() sdk.Coin
	AddTotalShares(amt sdk.Int)
	SubTotalShares(amt sdk.Int)
	GetPoolAsset(denom string) (PoolAsset, error)
	// UpdatePoolAssetBalance updates the balances for
	// the token with denomination coin.denom
	UpdatePoolAssetBalance(coin sdk.Coin) error
	// UpdatePoolAssetBalances calls UpdatePoolAssetBalance
	// on each constituent coin.
	UpdatePoolAssetBalances(coins sdk.Coins) error
	GetPoolAssets(denoms ...string) ([]PoolAsset, error)
	GetAllPoolAssets() []PoolAsset
	PokeTokenWeights(blockTime time.Time)
	GetTokenWeight(denom string) (sdk.Int, error)
	GetTokenBalance(denom string) (sdk.Int, error)
	NumAssets() int
	IsActive(curBlockTime time.Time) bool
}

var (
	MaxUserSpecifiedWeight    sdk.Int = sdk.NewIntFromUint64(1 << 20)
	GuaranteedWeightPrecision int64   = 1 << 30
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	key := append([]byte("pool"), sdk.Uint64ToBigEndian(poolId)...)
	return address.Module(ModuleName, key)
}

func ValidateUserSpecifiedWeight(weight sdk.Int) error {
	if !weight.IsPositive() {
		return sdkerrors.Wrap(ErrNotPositiveWeight, weight.String())
	}

	if weight.GTE(MaxUserSpecifiedWeight) {
		return sdkerrors.Wrap(ErrWeightTooLarge, weight.String())
	}
	return nil
}

func addSwapFee(pool PoolI, tokenAmountIn sdk.Dec) sdk.Dec {
	// tAI / (1-sf)
	return tokenAmountIn.Quo(sdk.OneDec().Sub(pool.GetPoolSwapFee()))
}

func subSwapFee(pool PoolI, tokenAmountIn sdk.Dec) sdk.Dec {
	// tAI * (1-sf)
	return tokenAmountIn.Mul(sdk.OneDec().Sub(pool.GetPoolSwapFee()))
}

func addSwapFeeWeightProportional(pool PoolI, tokenAmountIn, normalizedWeight sdk.Dec) sdk.Dec {
	// tAI / (1-(1-nw)*sf)
	return tokenAmountIn.Quo(sdk.OneDec().Sub(sdk.OneDec().Sub(normalizedWeight).Mul(pool.GetPoolSwapFee())))
}

func subSwapFeeWeightProportional(pool PoolI, tokenAmountIn, normalizedWeight sdk.Dec) sdk.Dec {
	// tAI * (1-(1-nw)*sf)
	return tokenAmountIn.Mul(sdk.OneDec().Sub(sdk.OneDec().Sub(normalizedWeight).Mul(pool.GetPoolSwapFee())))
}

func addExitFee(pool PoolI, poolAmountIn sdk.Dec) sdk.Dec {
	// pAI / (1-ef)
	return poolAmountIn.Quo(sdk.OneDec().Sub(pool.GetPoolExitFee()))
}

func subExitFee(pool PoolI, poolAmountIn sdk.Dec) sdk.Dec {
	// pAI * (1-ef)
	return poolAmountIn.Mul(sdk.OneDec().Sub(pool.GetPoolExitFee()))
}

// calcSpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool
// so spot_price = (B_in / W_in) / (B_out / W_out)
func CalcSpotPrice(
	pool PoolI,
	tokenIn, tokenOut string,
) (sdk.Dec, error) {
	assetIn, err := pool.GetPoolAsset(tokenIn)
	if err != nil {
		return sdk.Dec{}, err
	}
	assetOut, err := pool.GetPoolAsset(tokenOut)
	if err != nil {
		return sdk.Dec{}, err
	}

	number := assetIn.Token.Amount.ToDec().Quo(assetIn.Weight.ToDec())
	denom := assetOut.Token.Amount.ToDec().Quo(assetOut.Weight.ToDec())
	ratio := number.Quo(denom)

	return ratio, nil
}

// calcSpotPriceWithSwapFee returns the spot price of the pool accounting for
// the input taken by the swap fee.
// This is the weight-adjusted balance of the tokens in the pool.
// so spot_price = (B_in / W_in) / (B_out / W_out)
// and spot_price_with_fee = spot_price / (1 - swapfee)
func CalcSpotPriceWithSwapFee(
	pool PoolI,
	tokenIn, tokenOut string,
) (sdk.Dec, error) {
	spotPrice, err := CalcSpotPrice(pool, tokenIn, tokenOut)
	if err != nil {
		return sdk.Dec{}, err
	}
	// Q: why is this not just (1 - swapfee)
	// A: Its becasue its being applied to the other asset.
	// TODO: write this up more coherently
	// 1 / (1 - swapfee)
	scale := sdk.OneDec().Quo(sdk.OneDec().Sub(pool.GetPoolSwapFee()))

	return spotPrice.Mul(scale), nil
}

func getPoolInOutAssetsNormalized(pool PoolI, tokenIn, tokenOut string) (NormalizedPoolAsset, NormalizedPoolAsset, error) {
	assetIn, err := pool.GetPoolAsset(tokenIn)
	if err != nil {
		return NormalizedPoolAsset{}, NormalizedPoolAsset{}, err
	}
	assetOut, err := pool.GetPoolAsset(tokenOut)
	if err != nil {
		return NormalizedPoolAsset{}, NormalizedPoolAsset{}, err
	}
	totalWeight := pool.GetTotalWeight()
	return assetIn.Normalize(totalWeight), assetOut.Normalize(totalWeight), nil
}

func CalcOutGivenIn(
	pool PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
) (sdk.Dec, error) {
	assetIn, assetOut, err := getPoolInOutAssetsNormalized(pool, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInAmountFeeDeducted := subSwapFee(pool, tokenIn.Amount.ToDec())
	return pool.SolveConstantFunctionInvariant(
		assetIn.Token.Amount.ToDec(),
		assetIn.Weight,
		assetOut.Token.Amount.ToDec(),
		assetOut.Weight,
		tokenInAmountFeeDeducted,
	), nil
}

func CalcInGivenOut(
	pool PoolI,
	tokenOut sdk.Coin,
	tokenInDenom string,
) (sdk.Dec, error) {
	assetIn, assetOut, err := getPoolInOutAssetsNormalized(pool, tokenInDenom, tokenOut.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInAmountFeeDeducted := pool.SolveConstantFunctionInvariant(
		assetOut.Token.Amount.ToDec(),
		assetOut.Weight,
		assetIn.Token.Amount.ToDec(),
		assetIn.Weight,
		tokenOut.Amount.ToDec().Neg(),
	).Neg()
	fmt.Printf("%+v, %+v, %s\n", assetIn, assetOut, tokenInAmountFeeDeducted)
	return addSwapFee(pool, tokenInAmountFeeDeducted), nil
}

func CalcSingleInGivenPoolOut(
	pool PoolI,
	shareOutAmount sdk.Int,
	tokenInDenom string,
) (sdk.Dec, error) {
	asset, err := pool.GetPoolAsset(tokenInDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	normalized := asset.Normalize(pool.GetTotalWeight())
	tokenInAmountFeeDeducted := pool.SolveTokenFromShare(
		normalized.Token.Amount.ToDec(),
		normalized.Weight,
		pool.GetTotalShares().Amount.ToDec(),
		shareOutAmount.ToDec(),
	)
	return addSwapFeeWeightProportional(pool, tokenInAmountFeeDeducted, normalized.Weight), nil
}

func CalcSingleOutGivenPoolIn(
	pool PoolI,
	shareInAmount sdk.Int,
	tokenOutDenom string,
) (sdk.Dec, error) {
	asset, err := pool.GetPoolAsset(tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	normalized := asset.Normalize(pool.GetTotalWeight())

	shareInAmountExitFeeDeducted := subExitFee(pool, shareInAmount.ToDec())
	fmt.Printf("%+v, %+v, %s\n", pool, normalized, shareInAmountExitFeeDeducted.String())
	tokenOutAmount := pool.SolveTokenFromShare(
		normalized.Token.Amount.ToDec(),
		normalized.Weight,
		pool.GetTotalShares().Amount.ToDec(),
		shareInAmountExitFeeDeducted.Neg(),
	).Neg()
	tokenOutAmountFeeDeducted := subSwapFeeWeightProportional(pool, tokenOutAmount, normalized.Weight)
	fmt.Println(tokenOutAmountFeeDeducted)
	return tokenOutAmountFeeDeducted, nil
}

func CalcPoolInGivenSingleOut(
	pool PoolI,
	tokenOutFeeDeducted sdk.Coin,
) (sdk.Dec, error) {
	asset, err := pool.GetPoolAsset(tokenOutFeeDeducted.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	normalized := asset.Normalize(pool.GetTotalWeight())

	tokenOutAmount := addSwapFeeWeightProportional(pool, tokenOutFeeDeducted.Amount.ToDec(), normalized.Weight)
	shareInAmountFeeDeducted := pool.SolveShareFromToken(
		normalized.Token.Amount.ToDec(),
		normalized.Weight,
		pool.GetTotalShares().Amount.ToDec(),
		tokenOutAmount.Neg(),
	).Neg()
	return addExitFee(pool, shareInAmountFeeDeducted), nil
}

func CalcPoolOutGivenSingleIn(
	pool PoolI,
	tokenIn sdk.Coin,
) (sdk.Dec, error) {
	asset, err := pool.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	normalized := asset.Normalize(pool.GetTotalWeight())

	tokenInAmountFeeDeducted := subSwapFeeWeightProportional(pool, tokenIn.Amount.ToDec(), normalized.Weight)
	shareOutAmount := pool.SolveShareFromToken(
		normalized.Token.Amount.ToDec(),
		normalized.Weight,
		pool.GetTotalShares().Amount.ToDec(),
		tokenInAmountFeeDeducted,
	)
	return shareOutAmount, nil
}
