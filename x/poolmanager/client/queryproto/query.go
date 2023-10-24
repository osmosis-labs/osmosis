package queryproto

import (
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ZeroEstimateTradeBasedOnPriceImpactResponseOnErrInvalidMathApprox checks whether the error argument is
// a gammtypes.ErrInvalidMathApprox. This error is tolerated since nothing major went wrong, and we return
// a zero result in this case. Otherwise, we just forward the error, wrapped as an internal error.
func ZeroEstimateTradeBasedOnPriceImpactResponseOnErrInvalidMathApprox(
	req EstimateTradeBasedOnPriceImpactRequest, err error,
) (*EstimateTradeBasedOnPriceImpactResponse, error) {
	if errors.Is(err, gammtypes.ErrInvalidMathApprox) {
		return ZeroEstimateTradeBasedOnPriceImpactResponseFromRequest(req), nil
	}
	return nil, status.Error(codes.Internal, err.Error())
}

// ZeroEstimateTradeBasedOnPriceImpactResponseFromRequest is a helper function for creating a zero result
// based on the original denoms in a EstimateTradeBasedOnPriceImpactRequest.
func ZeroEstimateTradeBasedOnPriceImpactResponseFromRequest(
	req EstimateTradeBasedOnPriceImpactRequest,
) *EstimateTradeBasedOnPriceImpactResponse {
	return ZeroEstimateTradeBasedOnPriceImpactResponse(req.FromCoin.Denom, req.ToCoinDenom)
}

// ZeroEstimateTradeBasedOnPriceImpactResponse is a helper function for creating a zero result based on
// input coin denom and an output coin denom.
func ZeroEstimateTradeBasedOnPriceImpactResponse(
	inputCoinDenom, outputCoinDenom string,
) *EstimateTradeBasedOnPriceImpactResponse {
	return &EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  sdk.NewCoin(inputCoinDenom, math.ZeroInt()),
		OutputCoin: sdk.NewCoin(outputCoinDenom, math.ZeroInt()),
	}
}
