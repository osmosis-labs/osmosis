package concentrated_pool

import (
	"encoding/json"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.ConcentratedPoolExtension = &Pool{}
)

func NewConcentratedLiquidityPool(poolId uint64, denom0, denom1 string, currSqrtPrice sdk.Dec, currTick sdk.Int) (Pool, error) {
	denom0, denom1, err := types.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return Pool{}, err
	}
	pool := Pool{
		// TODO: move gammtypes.NewPoolAddress(poolId) to swaproutertypes
		Address:          gammtypes.NewPoolAddress(poolId).String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
		Liquidity:        sdk.ZeroDec(),
		Token0:           denom0,
		Token1:           denom1,
	}

	return pool, nil
}

// GetAddress returns the address of the concentrated liquidity pool
func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

// GetId returns the id of the concentrated liquidity pool
func (p Pool) GetId() uint64 {
	return p.Id
}

// String returns the json marshalled string of the pool
func (p Pool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// GetSwapFee returns the swap fee of the pool
func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}

// GetExitFee returns the exit fee of the pool
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}

// IsActive returns true if the pool is active
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// SpotPrice returns the spot price of the pool.
// If base asset is the Token0 of the pool, we use the current sqrt price of the pool.
// If not, we calculate the inverse of the current sqrt price of the pool.
func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	// if zero for one, we use the pool curr sqrt price directly.
	if p.Token0 == baseAssetDenom && p.Token1 == quoteAssetDenom {
		return p.CurrentSqrtPrice.Power(2), nil
	} else if p.Token1 == baseAssetDenom && p.Token0 == quoteAssetDenom { // if not, we calculate the reverse spot price by 1 / currentSqrtPrice^2
		return sdk.NewDec(1).Quo(p.CurrentSqrtPrice.Power(2)), nil
	} else {
		return sdk.Dec{}, fmt.Errorf("base asset denom %s is not in the pool", baseAssetDenom)
	}
}

// GetTotalShares returns the total shares of the pool
func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
}

// GetToken0 returns the token0 of the pool
func (p Pool) GetToken0() string {
	return p.Token0
}

// GetToken1 returns the token1 of the pool
func (p Pool) GetToken1() string {
	return p.Token1
}

// GetCurrentSqrtPrice returns the current sqrt price of the pool
func (p Pool) GetCurrentSqrtPrice() sdk.Dec {
	return p.CurrentSqrtPrice
}

// GetCurrentTick returns the current tick of the pool
func (p Pool) GetCurrentTick() sdk.Int {
	return p.CurrentTick
}

// GetLiquidity returns the liquidity of the pool
func (p Pool) GetLiquidity() sdk.Dec {
	return p.Liquidity
}

// UpdateLiquidity updates the liquidity of the pool. Note that this method is mutative.
func (p *Pool) UpdateLiquidity(newLiquidity sdk.Dec) {
	p.Liquidity = p.Liquidity.Add(newLiquidity)
}
