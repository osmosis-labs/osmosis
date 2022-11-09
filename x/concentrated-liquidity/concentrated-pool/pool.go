package concentrated_pool

import (
	"encoding/json"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.PoolI = &Pool{}
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

func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

func (p Pool) GetId() uint64 {
	return p.Id
}

func (p Pool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}

func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	// if zero for one, we use the pool curr sqrt price directly.
	if p.Token0 == baseAssetDenom {
		return p.CurrentSqrtPrice.Power(2), nil
	} else { // if not, we calculate the reverse spot price by 1 / currentSqrtPrice^2
		return sdk.NewDec(1).Quo(p.CurrentSqrtPrice.Power(2)), nil
	}
}

func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
}

func (p Pool) GetToken0() string {
	return p.Token0
}

func (p Pool) GetToken1() string {
	return p.Token1
}

func (p Pool) GetCurrentSqrtPrice() sdk.Dec {
	return p.CurrentSqrtPrice
}

func (p Pool) GetCurrentTick() sdk.Int {
	return p.CurrentTick
}
func (p Pool) GetLiquidity() sdk.Dec {
	return p.Liquidity
}

func (p *Pool) UpdateLiquidity(newLiquidity sdk.Dec) {
	p.Liquidity = p.Liquidity.Add(newLiquidity)
}
