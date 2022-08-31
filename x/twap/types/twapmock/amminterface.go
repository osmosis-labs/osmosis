package twapmock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

type ProgrammedAmmInterface struct {
	underlyingKeeper     types.AmmInterface
	programmedSpotPrice  map[spotPriceInput]spotPriceResult
	programmedPoolDenoms map[uint64]poolDenomResponse
}

// TODO, generalize to do a sum type on denoms
type spotPriceInput struct {
	poolId     uint64
	baseDenom  string
	quoteDenom string
}
type spotPriceResult struct {
	sp  sdk.Dec
	err error
}
type poolDenomResponse struct {
	denoms []string
	err    error
}

func NewProgrammedAmmInterface(underlyingKeeper types.AmmInterface) *ProgrammedAmmInterface {
	return &ProgrammedAmmInterface{underlyingKeeper: underlyingKeeper}
}

func (p *ProgrammedAmmInterface) ProgramPoolDenomsOverride(poolId uint64, overrideDenoms []string, overrideErr error) {
	p.programmedPoolDenoms[poolId] = poolDenomResponse{overrideDenoms, overrideErr}
}

func (p *ProgrammedAmmInterface) ProgramPoolSpotPriceOverride(poolId uint64,
	baseDenom, quoteDenom string, overrideSp sdk.Dec, overrideErr error) {
	input := spotPriceInput{poolId, baseDenom, quoteDenom}
	p.programmedSpotPrice[input] = spotPriceResult{overrideSp, overrideErr}
}

func (s *ProgrammedAmmInterface) GetPoolDenoms(ctx sdk.Context, poolId uint64) (denoms []string, err error) {
	if res, ok := s.programmedPoolDenoms[poolId]; ok {
		return res.denoms, res.err
	}
	return s.underlyingKeeper.GetPoolDenoms(ctx, poolId)
}

func (s *ProgrammedAmmInterface) CalculateSpotPrice(ctx sdk.Context,
	poolId uint64,
	baseDenom,
	quoteDenom string) (price sdk.Dec, err error) {
	input := spotPriceInput{poolId, baseDenom, quoteDenom}
	if res, ok := s.programmedSpotPrice[input]; ok {
		return res.sp, res.err
	}
	return s.underlyingKeeper.CalculateSpotPrice(ctx, poolId, baseDenom, quoteDenom)
}

var _ types.AmmInterface = &ProgrammedAmmInterface{}
