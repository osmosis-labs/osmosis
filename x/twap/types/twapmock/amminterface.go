package twapmock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/twap/types"
)

var _ types.AmmInterface = &ProgrammedAmmInterface{}

type ProgrammedAmmInterface struct {
	underlyingKeeper     types.AmmInterface
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
	Sp  sdk.Dec
	Err error
}

type poolDenomsResult struct {
	poolDenoms map[string]struct{}
	err        error
}

func NewProgrammedAmmInterface(underlyingKeeper types.AmmInterface) *ProgrammedAmmInterface {
	return &ProgrammedAmmInterface{
		underlyingKeeper:     underlyingKeeper,
		programmedSpotPrice:  map[SpotPriceInput]SpotPriceResult{},
		programmedPoolDenoms: map[uint64]poolDenomsResult{},
	}
}

func (p *ProgrammedAmmInterface) ProgramPoolDenomsOverride(poolId uint64, overrideDenoms []string, overrideErr error) {
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

func (p *ProgrammedAmmInterface) ProgramPoolSpotPriceOverride(poolId uint64,
	quoteDenom, baseDenom string, overrideSp sdk.Dec, overrideErr error,
) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	p.programmedSpotPrice[input] = SpotPriceResult{overrideSp, overrideErr}
}

func (p *ProgrammedAmmInterface) GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error) {
	if res, ok := p.programmedPoolDenoms[poolId]; ok {
		result := make([]string, 0, len(res.poolDenoms))
		for denom := range res.poolDenoms {
			result = append(result, denom)
		}
		return result, res.err
	}
	return p.underlyingKeeper.GetPoolDenoms(ctx, poolId)
}

func (p *ProgrammedAmmInterface) CalculateSpotPrice(ctx sdk.Context,
	poolId uint64,
	quoteDenom,
	baseDenom string,
) (price sdk.Dec, err error) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	if res, ok := p.programmedSpotPrice[input]; ok {
		return res.Sp, res.Err
	}
	return p.underlyingKeeper.CalculateSpotPrice(ctx, poolId, quoteDenom, baseDenom)
}
