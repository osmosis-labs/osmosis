package queryproto

import (
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ZeroEstimateTradeBasedOnPriceImpactResponseOnErrInvalidMathApprox(
	req EstimateTradeBasedOnPriceImpactRequest, err error,
) (*EstimateTradeBasedOnPriceImpactResponse, error) {
	if errors.Is(err, gammtypes.ErrInvalidMathApprox) {
		return ZeroEstimateTradeBasedOnPriceImpactResponseFromRequest(req), nil
	}
	return nil, status.Error(codes.Internal, err.Error())
}

func ZeroEstimateTradeBasedOnPriceImpactResponseFromRequest(
	req EstimateTradeBasedOnPriceImpactRequest,
) *EstimateTradeBasedOnPriceImpactResponse {
	return ZeroEstimateTradeBasedOnPriceImpactResponse(req.FromCoin.Denom, req.ToCoinDenom)
}

func ZeroEstimateTradeBasedOnPriceImpactResponse(
	inputCoinDenom, outputCoinDenom string,
) *EstimateTradeBasedOnPriceImpactResponse {
	return &EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  sdk.NewCoin(inputCoinDenom, math.ZeroInt()),
		OutputCoin: sdk.NewCoin(outputCoinDenom, math.ZeroInt()),
	}
}
