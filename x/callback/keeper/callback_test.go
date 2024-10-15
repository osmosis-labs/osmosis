package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func (s *KeeperTestSuite) TestSaveCallback() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAddr3 := contractAddresses[2]
	contractAdminAcc := s.chain.GetAccount(0)
	notContractAdminAcc := s.chain.GetAccount(1)
	// contractOwnerAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)

	// Setting callback module as contract owner
	blockedModuleAddr := s.chain.GetApp().AccountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
	s.Require().True(s.chain.GetApp().BankKeeper.BlockedAddr(blockedModuleAddr))

	params, err := keeper.GetParams(ctx)
	s.Require().NoError(err)

	testCases := []struct {
		testCase    string
		callback    types.Callback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: contract address is invalid",
			callback: types.Callback{
				ContractAddress: "👻",
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: contract does not exist",
			callback: types.Callback{
				ContractAddress: contractAddr3.String(),
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "FAIL: sender not authorized to modify",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      notContractAdminAcc.Address.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "FAIL: callback height is in the past",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  99,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrCallbackHeightNotInFuture,
		},
		{
			testCase: "FAIL: callback height is current height",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  ctx.BlockHeight(),
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrCallbackHeightNotInFuture,
		},
		{
			testCase: "FAIL: callback is too far in the future",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  ctx.BlockHeight() + int64(params.MaxFutureReservationLimit) + 1,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrCallbackHeightTooFarInFuture,
		},
		{
			testCase: "FAIL: sender is a blocked address",
			callback: types.Callback{
				ContractAddress: contractAddr2.String(),
				JobId:           1,
				CallbackHeight:  102,
				ReservedBy:      blockedModuleAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "OK: save callback - sender is contract",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: false,
		},

		{
			testCase: "OK: save callback - sender is contract admin",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           3,
				CallbackHeight:  101,
				ReservedBy:      contractAdminAcc.Address.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: false,
		},
		{
			testCase: "OK: save callback - sender is contract admin but address is in uppercase",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           5,
				CallbackHeight:  101,
				ReservedBy:      strings.ToUpper(contractAdminAcc.Address.String()),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: false,
		},
		{
			testCase: "FAIL: callback already exists",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrCallbackExists,
		},
		{
			testCase: "FAIL: block is filled with max number of callbacks",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           4,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			expectError: true,
			errorType:   types.ErrBlockFilled,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			err := keeper.SaveCallback(ctx, tc.callback)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorContains(err, tc.errorType.Error())
			} else {
				s.Require().NoError(err)
				// Ensuring the callback exists now
				exists, err := keeper.ExistsCallback(ctx, tc.callback.CallbackHeight, tc.callback.ContractAddress, tc.callback.JobId)
				s.Require().NoError(err)
				s.Require().True(exists)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDeleteCallback() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	notContractAdminAcc := s.chain.GetAccount(1)
	// contractOwnerAcc := s.chain.GetAccount(2)

	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	// Same contract requesting callback at same height with diff job id
	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  101,
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

	testCases := []struct {
		testCase    string
		callback    types.Callback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: Invalid contract address",
			callback: types.Callback{
				ContractAddress: "👻",
				JobId:           0,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: Not authorized to delete callback",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           0,
				CallbackHeight:  101,
				ReservedBy:      notContractAdminAcc.Address.String(),
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "FAIL: Callback does not exist",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           0,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
			},
			expectError: false, // Should silently fail. MsgSrvr ensures that callback exists before calling keeper
		},
		{
			testCase: "OK: Success delete - sender is contract",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  101,
				ReservedBy:      contractAddr.String(),
			},
			expectError: false,
		},
		{
			testCase: "OK: Success delete - sender is contract admin",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				JobId:           2,
				CallbackHeight:  101,
				ReservedBy:      contractAdminAcc.Address.String(),
			},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			err := keeper.DeleteCallback(ctx, tc.callback.ReservedBy, tc.callback)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorContains(err, tc.errorType.Error())
			} else {
				s.Require().NoError(err)
				// Ensuring the callback does not exist anymore
				exists, err := keeper.ExistsCallback(ctx, tc.callback.CallbackHeight, tc.callback.ContractAddress, tc.callback.JobId)
				s.Require().NoError(err)
				s.Require().False(exists)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetCallbacksByHeight() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	callbackHeight := int64(101)

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

	s.Run("OK: Get all three existing callbacks at height 101", func() {
		callbacks, err := keeper.GetCallbacksByHeight(ctx, callbackHeight)
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(callbacks))
	})
	s.Run("OK: Get zero existing callbacks at height 102", func() {
		callbacks, err := keeper.GetCallbacksByHeight(ctx, callbackHeight+1)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(callbacks))
	})
}

func (s *KeeperTestSuite) TestGetAllCallbacks() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	callbackHeight := int64(105)

	s.Run("OK: Get zero existing callbacks", func() {
		callbacks, err := keeper.GetAllCallbacks(ctx)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(callbacks))
	})

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
	callback.CallbackHeight = callbackHeight + 1
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	callback.CallbackHeight = callbackHeight + 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	s.Run("OK: Get all existing callbacks - 3", func() {
		callbacks, err := keeper.GetAllCallbacks(ctx)
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(callbacks))
	})
}

func (s *KeeperTestSuite) TestIterateCallbacksByHeight() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	callbackHeight := int64(101)

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

	s.Run("OK: Get all three existing callbacks at height 101", func() {
		count := 0
		keeper.IterateCallbacksByHeight(ctx, callbackHeight, func(callback types.Callback) bool {
			count++
			return false
		})
		s.Assert().Equal(3, count)
	})

	s.Run("OK: Get one existing callbacks at height 102", func() {
		count := 0
		keeper.IterateCallbacksByHeight(ctx, callbackHeight+1, func(callback types.Callback) bool {
			count++
			return false
		})
		s.Assert().Equal(1, count)
	})
}
