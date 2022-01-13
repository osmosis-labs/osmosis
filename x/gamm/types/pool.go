package types

import (
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

	Swap() SwapI

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
	GetPoolAssets(denoms ...string) ([]PoolAsset, error)
	GetAllPoolAssets() []PoolAsset
	// UpdatePoolAssetBalance updates the balances for
	// the token amount difference with denomination coin.denom
	AddPoolAssetBalance(coins ...sdk.Coin) error
	SubPoolAssetBalance(coins ...sdk.Coin) error
	PokeTokenWeights(blockTime time.Time)
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

func addSwapFee(tokenAmountIn, swapFee sdk.Dec) sdk.Dec {
	// tAI / (1-sf)
	return tokenAmountIn.Quo(sdk.OneDec().Sub(swapFee))
}

func subSwapFee(tokenAmountIn, swapFee sdk.Dec) sdk.Dec {
	// tAI * (1-sf)
	return tokenAmountIn.Mul(sdk.OneDec().Sub(swapFee))
}

func addSwapFeeWeightProportional(tokenAmountIn, normalizedWeight, swapFee sdk.Dec) sdk.Dec {
	// tAI / (1-(1-nw)*sf)
	return tokenAmountIn.Quo(sdk.OneDec().Sub(sdk.OneDec().Sub(normalizedWeight).Mul(swapFee)))
}

func subSwapFeeWeightProportional(tokenAmountIn, normalizedWeight, swapFee sdk.Dec) sdk.Dec {
	// tAI * (1-(1-nw)*sf)
	return tokenAmountIn.Mul(sdk.OneDec().Sub(sdk.OneDec().Sub(normalizedWeight).Mul(swapFee)))
}

func addExitFee(poolAmountIn, exitFee sdk.Dec) sdk.Dec {
	// pAI / (1-ef)
	return poolAmountIn.Quo(sdk.OneDec().Sub(exitFee))
}

func subExitFee(poolAmountIn, exitFee sdk.Dec) sdk.Dec {
	// pAI * (1-ef)
	return poolAmountIn.Mul(sdk.OneDec().Sub(exitFee))
}

// calcSpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool
// so spot_price = (B_in / W_in) / (B_out / W_out)
func CalcSpotPrice(
	assetIn, assetOut PoolAsset,
) sdk.Dec {
	number := assetIn.Token.Amount.ToDec().Quo(assetIn.Weight.ToDec())
	denom := assetOut.Token.Amount.ToDec().Quo(assetOut.Weight.ToDec())
	ratio := number.Quo(denom)

	return ratio
}

// calcSpotPriceWithSwapFee returns the spot price of the pool accounting for
// the input taken by the swap fee.
// This is the weight-adjusted balance of the tokens in the pool.
// so spot_price = (B_in / W_in) / (B_out / W_out)
// and spot_price_with_fee = spot_price / (1 - swapfee)
func CalcSpotPriceWithSwapFee(
	assetIn, assetOut PoolAsset,
	swapFee sdk.Dec,
) sdk.Dec {
	spotPrice := CalcSpotPrice(assetIn, assetOut)
	// Q: why is this not just (1 - swapfee)
	// A: Its becasue its being applied to the other asset.
	// TODO: write this up more coherently
	// 1 / (1 - swapfee)
	scale := sdk.OneDec().Quo(sdk.OneDec().Sub(swapFee))

	return spotPrice.Mul(scale)
}

func CalcOutGivenIn(
	swap SwapI,
	assetIn, assetOut NormalizedPoolAsset,
	tokenInAmount sdk.Int,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenInAmountFeeDeducted := subSwapFee(tokenInAmount.ToDec(), swapFee)
	return swap.SolveConstantFunctionInvariant(
		assetIn.Token.Amount.ToDec(),
		assetIn.Weight,
		assetOut.Token.Amount.ToDec(),
		assetOut.Weight,
		tokenInAmountFeeDeducted,
	)
}

func CalcInGivenOut(
	swap SwapI,
	assetIn, assetOut NormalizedPoolAsset,
	tokenOutAmount sdk.Int,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenInAmountFeeDeducted := swap.SolveConstantFunctionInvariant(
		assetOut.Token.Amount.ToDec(),
		assetOut.Weight,
		assetIn.Token.Amount.ToDec(),
		assetIn.Weight,
		tokenOutAmount.ToDec().Neg(),
	).Neg()
	return addSwapFee(tokenInAmountFeeDeducted, swapFee)
}

func CalcSingleInGivenPoolOut(
	swap SwapI,
	asset NormalizedPoolAsset,
	totalShares sdk.Int,
	shareOutAmount sdk.Int,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenInAmountFeeDeducted := swap.SolveTokenFromShare(
		asset.Token.Amount.ToDec(),
		asset.Weight,
		totalShares.ToDec(),
		shareOutAmount.ToDec(),
	)
	return addSwapFeeWeightProportional(tokenInAmountFeeDeducted, asset.Weight, swapFee)
}

func CalcSingleOutGivenPoolIn(
	swap SwapI,
	asset NormalizedPoolAsset,
	totalShares sdk.Int,
	shareInAmount sdk.Int,
	swapFee, exitFee sdk.Dec,
) sdk.Dec {
	shareInAmountExitFeeDeducted := subExitFee(shareInAmount.ToDec(), exitFee)
	tokenOutAmount := swap.SolveTokenFromShare(
		asset.Token.Amount.ToDec(),
		asset.Weight,
		totalShares.ToDec(),
		shareInAmountExitFeeDeducted.Neg(),
	).Neg()
	tokenOutAmountFeeDeducted := subSwapFeeWeightProportional(tokenOutAmount, asset.Weight, swapFee)
	return tokenOutAmountFeeDeducted
}

func CalcPoolInGivenSingleOut(
	swap SwapI,
	asset NormalizedPoolAsset,
	totalShares sdk.Int,
	tokenOutFeeDeducted sdk.Int,
	swapFee, exitFee sdk.Dec,
) sdk.Dec {
	tokenOutAmount := addSwapFeeWeightProportional(tokenOutFeeDeducted.ToDec(), asset.Weight, swapFee)
	shareInAmountFeeDeducted := swap.SolveShareFromToken(
		asset.Token.Amount.ToDec(),
		asset.Weight,
		totalShares.ToDec(),
		tokenOutAmount.Neg(),
	).Neg()
	return addExitFee(shareInAmountFeeDeducted, exitFee)
}

func CalcPoolOutGivenSingleIn(
	swap SwapI,
	asset NormalizedPoolAsset,
	totalShares sdk.Int,
	tokenInAmount sdk.Int,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenInAmountFeeDeducted := subSwapFeeWeightProportional(tokenInAmount.ToDec(), asset.Weight, swapFee)
	shareOutAmount := swap.SolveShareFromToken(
		asset.Token.Amount.ToDec(),
		asset.Weight,
		totalShares.ToDec(),
		tokenInAmountFeeDeducted,
	)
	return shareOutAmount
}

func CalcMultiGivenPool(
	assets []PoolAsset,
	totalSharesAmount sdk.Int,
	shareAmount sdk.Int,
) (sdk.Coins, error) {
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	// (tokens per share) * number of shares out = # tokens * (# shares out / cur total shares)
	shareRatio := shareAmount.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return nil, sdkerrors.Wrapf(ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	poolAssetsDiff := make([]sdk.Coin, 0, len(assets))
	// Transfer the PoolAssets tokens to the pool's module account from the user account.
	for _, asset := range assets {
		tokenDiffAmount := shareRatio.MulInt(asset.Token.Amount).TruncateInt()
		if tokenDiffAmount.LTE(sdk.ZeroInt()) {
			return nil, sdkerrors.Wrapf(ErrInvalidMathApprox, "token amount is zero or negative")
		}

		poolAssetsDiff = append(poolAssetsDiff, sdk.NewCoin(asset.Token.Denom, tokenDiffAmount))
	}

	return poolAssetsDiff, nil
}

func CalcJoin(
	assets []PoolAsset,
	totalSharesAmount sdk.Int,
	shareOutAmount sdk.Int,
) (sdk.Coins, error) {
	return CalcMultiGivenPool(assets, totalSharesAmount, shareOutAmount)
}

func CalcExit(
	assets []PoolAsset,
	totalSharesAmount sdk.Int,
	shareInAmount sdk.Int,
	exitFee sdk.Dec,
) (sdk.Coins, error) {
	shareInAmountAfterExitFee := shareInAmount.Sub(exitFee.MulInt(shareInAmount).TruncateInt())
	return CalcMultiGivenPool(assets, totalSharesAmount, shareInAmountAfterExitFee)
}
