package balancer

import (
	"errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

func NewPoolParams(spreadFactor, exitFee osmomath.Dec, params *SmoothWeightChangeParams) PoolParams {
	return PoolParams{
		SwapFee:                  spreadFactor,
		ExitFee:                  exitFee,
		SmoothWeightChangeParams: params,
	}
}

func (params PoolParams) Validate(poolWeights []PoolAsset) error {
	if params.ExitFee.IsNegative() {
		return types.ErrNegativeExitFee
	}

	if params.ExitFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchExitFee
	}

	if params.SwapFee.IsNegative() {
		return types.ErrNegativeSpreadFactor
	}

	if params.SwapFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchSpreadFactor
	}

	if params.SmoothWeightChangeParams != nil {
		targetWeights := params.SmoothWeightChangeParams.TargetPoolWeights
		// Ensure it has the right number of weights
		if len(targetWeights) != len(poolWeights) {
			return types.ErrPoolParamsInvalidNumDenoms
		}
		// Validate all user specified weights
		for _, v := range targetWeights {
			err := ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
		}
		// Ensure that all the target weight denoms are same as pool asset weights
		sortedTargetPoolWeights := sortPoolAssetsOutOfPlaceByDenom(targetWeights)
		sortedPoolWeights := sortPoolAssetsOutOfPlaceByDenom(poolWeights)
		for i, v := range sortedPoolWeights {
			if sortedTargetPoolWeights[i].Token.Denom != v.Token.Denom {
				return types.ErrPoolParamsInvalidDenom
			}
		}

		// No start time validation needed

		// We do not need to validate InitialPoolWeights, as we set that ourselves
		// in setInitialPoolParams

		// TODO: Is there anything else we can validate for duration?
		if params.SmoothWeightChangeParams.Duration <= 0 {
			return errors.New("params.SmoothWeightChangeParams must have a positive duration")
		}
	}

	return nil
}

func (params PoolParams) GetPoolSpreadFactor() osmomath.Dec {
	return params.SwapFee
}

func (params PoolParams) GetPoolExitFee() osmomath.Dec {
	return params.ExitFee
}
