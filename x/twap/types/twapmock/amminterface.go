package twapmock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

var _ types.PoolManagerInterface = &ProgrammedPoolManagerInterface{}

type ProgrammedPoolManagerInterface struct {
	underlyingKeeper     types.PoolManagerInterface
	programmedSpotPrice  map[SpotPriceInput]SpotPriceResult
	programmedPoolDenoms map[uint64]poolDenomsResult
}

// TODO, generalize to do a sum type on denoms
type SpotPriceInput struct {
	poolId     uint64
	baseDenom  string
	quoteDenom string
}
type SpotPriceResult struct {
	Sp  osmomath.Dec
	Err error
}

type poolDenomsResult struct {
	poolDenoms map[string]struct{}
	err        error
}

func NewProgrammedAmmInterface(underlyingKeeper types.PoolManagerInterface) *ProgrammedPoolManagerInterface {
	return &ProgrammedPoolManagerInterface{
		underlyingKeeper:     underlyingKeeper,
		programmedSpotPrice:  map[SpotPriceInput]SpotPriceResult{},
		programmedPoolDenoms: map[uint64]poolDenomsResult{},
	}
}

func (p *ProgrammedPoolManagerInterface) ProgramPoolDenomsOverride(poolId uint64, overrideDenoms []string, overrideErr error) {
	var poolDenoms map[string]struct{}
	if existingForPool, ok := p.programmedPoolDenoms[poolId]; ok {
		poolDenoms = existingForPool.poolDenoms
	} else {
		poolDenoms = map[string]struct{}{}
	}
	for _, denom := range overrideDenoms {
		poolDenoms[denom] = struct{}{}
	}
	p.programmedPoolDenoms[poolId] = poolDenomsResult{poolDenoms, overrideErr}
}

func (p *ProgrammedPoolManagerInterface) ProgramPoolSpotPriceOverride(poolId uint64,
	quoteDenom, baseDenom string, overrideSp osmomath.Dec, overrideErr error,
) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	p.programmedSpotPrice[input] = SpotPriceResult{overrideSp, overrideErr}
}

func (p *ProgrammedPoolManagerInterface) RouteGetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error) {
	if res, ok := p.programmedPoolDenoms[poolId]; ok {
		result := make([]string, 0, len(res.poolDenoms))
		for denom := range res.poolDenoms {
			result = append(result, denom)
		}
		return result, res.err
	}
	return p.underlyingKeeper.RouteGetPoolDenoms(ctx, poolId)
}

func (p *ProgrammedPoolManagerInterface) RouteCalculateSpotPrice(ctx sdk.Context,
	poolId uint64,
	quoteDenom,
	baseDenom string,
) (price osmomath.BigDec, err error) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	if res, ok := p.programmedSpotPrice[input]; ok {
		if (res.Sp == osmomath.Dec{}) {
			return osmomath.BigDec{}, res.Err
		}

		return osmomath.BigDecFromDec(res.Sp), res.Err
	}
	return p.underlyingKeeper.RouteCalculateSpotPrice(ctx, poolId, quoteDenom, baseDenom)
}

func (p *ProgrammedPoolManagerInterface) GetNextPoolId(ctx sdk.Context) uint64 {
	return p.underlyingKeeper.GetNextPoolId(ctx)
}
