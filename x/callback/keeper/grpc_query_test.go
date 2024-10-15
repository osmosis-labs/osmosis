package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"
	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	callbackKeeper "github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func (s *KeeperTestSuite) TestCallbacks() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(101), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	callbackHeight := int64(102)
	// Same contract requesting callback at same height with diff job id
	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  callbackHeight,
		ReservedBy:      contractAddr.String(),
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &validCoin,
			BlockReservationFees:  &validCoin,
			FutureReservationFees: &validCoin,
			SurplusFees:           &validCoin,
		},
	}
	err := keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)
	callback.JobId = 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)
	callback.JobId = 3
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)
	// Same contract requesting callback at diff height with diff job id
	callback.JobId = 4
	callback.CallbackHeight = callbackHeight + 1
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	queryServer := callbackKeeper.NewQueryServer(keeper)

	testCases := []struct {
		testCase             string
		input                func() *types.QueryCallbacksRequest
		expectError          bool
		noOfCallbackExpected int
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.QueryCallbacksRequest {
				return nil
			},
			expectError:          true,
			noOfCallbackExpected: 0,
		},
		{
			testCase: "OK: no callbacks at requested height",
			input: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: 100,
				}
			},
			expectError:          false,
			noOfCallbackExpected: 0,
		},
		{
			testCase: "OK: get callbacks at requested height. there are three callbacks",
			input: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: callbackHeight,
				}
			},
			expectError:          false,
			noOfCallbackExpected: 3,
		},
		{
			testCase: "OK: get callbacks at requested height. there is one callback",
			input: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: callbackHeight + 1,
				}
			},
			expectError:          false,
			noOfCallbackExpected: 1,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := queryServer.Callbacks(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.noOfCallbackExpected, len(res.Callbacks))
			}
		})
	}
}

func (s *KeeperTestSuite) TestEstimateCallbackFees() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(101), s.chain.GetApp().CallbackKeeper
	queryServer := callbackKeeper.NewQueryServer(keeper)
	zeroCoin := sdk.NewInt64Coin("stake", 0)

	// Setting up custom params where the reservation multipliers are 0
	// and the max callback gas limit is 1, so tx fee is same as computational price of gas
	params, err := keeper.GetParams(ctx)
	s.Require().NoError(err)
	err = keeper.SetParams(ctx, types.Params{
		CallbackGasLimit:               1,
		MaxBlockReservationLimit:       params.MaxBlockReservationLimit,
		MaxFutureReservationLimit:      params.MaxFutureReservationLimit,
		FutureReservationFeeMultiplier: math.LegacyMustNewDecFromStr("0"),
		BlockReservationFeeMultiplier:  math.LegacyMustNewDecFromStr("0"),
		MinPriceOfGas:                  params.MinPriceOfGas,
	})
	s.Require().NoError(err)
	expectedTxFeeAmount := params.GetMinPriceOfGas().Amount
	expectedTxFeeCoin := sdk.NewInt64Coin("stake", expectedTxFeeAmount.Int64())

	testCases := []struct {
		testCase       string
		input          func() *types.QueryEstimateCallbackFeesRequest
		expectError    bool
		expectedOutput *types.QueryEstimateCallbackFeesResponse
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.QueryEstimateCallbackFeesRequest {
				return nil
			},
			expectError:    true,
			expectedOutput: nil,
		},
		{
			testCase: "FAIL: height is in the past",
			input: func() *types.QueryEstimateCallbackFeesRequest {
				return &types.QueryEstimateCallbackFeesRequest{
					BlockHeight: 100,
				}
			},
			expectError:    true,
			expectedOutput: nil,
		},
		{
			testCase: "FAIL: height is the current block height",
			input: func() *types.QueryEstimateCallbackFeesRequest {
				return &types.QueryEstimateCallbackFeesRequest{
					BlockHeight: 101,
				}
			},
			expectError:    true,
			expectedOutput: nil,
		},
		{
			testCase: "OK: fetch fees for next height",
			input: func() *types.QueryEstimateCallbackFeesRequest {
				return &types.QueryEstimateCallbackFeesRequest{
					BlockHeight: 102,
				}
			},
			expectError: false,
			expectedOutput: &types.QueryEstimateCallbackFeesResponse{
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &expectedTxFeeCoin,
					BlockReservationFees:  &zeroCoin,
					FutureReservationFees: &zeroCoin,
				},
				TotalFees: &expectedTxFeeCoin,
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := queryServer.EstimateCallbackFees(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput.TotalFees, res.TotalFees)
				s.Require().Equal(tc.expectedOutput.FeeSplit.BlockReservationFees, res.FeeSplit.BlockReservationFees)
				s.Require().Equal(tc.expectedOutput.FeeSplit.FutureReservationFees, res.FeeSplit.FutureReservationFees)
				s.Require().Equal(tc.expectedOutput.FeeSplit.TransactionFees, res.FeeSplit.TransactionFees)
				s.Require().Equal(tc.expectedOutput.FeeSplit.SurplusFees, res.FeeSplit.SurplusFees)
			}
		})
	}
}
