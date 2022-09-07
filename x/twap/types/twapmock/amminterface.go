package twapmock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

var _ types.AmmInterface = &ProgrammedAmmInterface{}

type ProgrammedAmmInterface struct {
	underlyingKeeper     types.AmmInterface
	programmedSpotPrice  map[SpotPriceInput]SpotPriceResult
	programmedPoolDenoms map[uint64]poolDenomResponse
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
type poolDenomResponse struct {
	denoms []string
	err    error
}

func NewProgrammedAmmInterface(underlyingKeeper types.AmmInterface) *ProgrammedAmmInterface {
	return &ProgrammedAmmInterface{
		underlyingKeeper:     underlyingKeeper,
		programmedSpotPrice:  map[SpotPriceInput]SpotPriceResult{},
		programmedPoolDenoms: map[uint64]poolDenomResponse{},
	}
}

func (p *ProgrammedAmmInterface) ProgramPoolDenomsOverride(poolId uint64, overrideDenoms []string, overrideErr error) {
	p.programmedPoolDenoms[poolId] = poolDenomResponse{overrideDenoms, overrideErr}
}

func (p *ProgrammedAmmInterface) ProgramPoolSpotPriceOverride(poolId uint64,
	baseDenom, quoteDenom string, overrideSp sdk.Dec, overrideErr error) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	p.programmedSpotPrice[input] = SpotPriceResult{overrideSp, overrideErr}
}

func (p *ProgrammedAmmInterface) GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error) {
	if res, ok := p.programmedPoolDenoms[poolId]; ok {
		return res.denoms, res.err
	}
	return p.underlyingKeeper.GetPoolDenoms(ctx, poolId)
}

func (p *ProgrammedAmmInterface) CalculateSpotPrice(ctx sdk.Context,
	poolId uint64,
	baseDenom,
	quoteDenom string) (price sdk.Dec, err error) {
	input := SpotPriceInput{poolId, baseDenom, quoteDenom}
	if res, ok := p.programmedSpotPrice[input]; ok {
		return res.Sp, res.Err
	}
	return p.underlyingKeeper.CalculateSpotPrice(ctx, poolId, baseDenom, quoteDenom)
}
