package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	callbackKeeper "github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func (s *KeeperTestSuite) TestRequestCallback() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(101), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractAdminBalance := s.chain.GetBalance(contractAdminAcc.Address)

	msgServer := callbackKeeper.NewMsgServer(keeper)

	testCases := []struct {
		testCase    string
		input       func() *types.MsgRequestCallback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgRequestCallback {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "OK: successfully register callback with 0 fees",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  102,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 0),
				}
			},
			expectError: false,
			errorType:   nil,
		},
		{
			testCase: "FAIL: contract does not exist",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAdminAcc.Address.String(),
					JobId:           1,
					CallbackHeight:  102,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 100000000),
				}
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "OK: successfully register callback",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  120,
					Sender:          contractAdminAcc.Address.String(),
					Fees:            sdk.NewInt64Coin("stake", 100000000),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := msgServer.RequestCallback(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorIs(err, tc.errorType)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(&types.MsgRequestCallbackResponse{}, res)
				// Ensuring the callback exists now
				exists, err := keeper.ExistsCallback(ctx, req.CallbackHeight, req.ContractAddress, req.JobId)
				s.Require().NoError(err)
				s.Require().True(exists)
				// Ensure account balance has been updated
				contractAdminBalance = contractAdminBalance.Sub(req.Fees)
				if req.Sender == contractAddr.String() {
					s.Require().True(contractAdminBalance.Equal(s.chain.GetBalance(contractAdminAcc.Address)))
				} else {
					s.Require().True(contractAdminBalance.Equal(s.chain.GetBalance(sdk.MustAccAddressFromBech32(req.Sender))))
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestCancelCallback() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(102), s.chain.GetApp().CallbackKeeper
	contractViewer := e2eTesting.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	msgServer := callbackKeeper.NewMsgServer(keeper)
	// Setting up an existing callback to delete
	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  130,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            sdk.NewInt64Coin("stake", 100),
	}
	_, err := msgServer.RequestCallback(ctx, reqMsg)
	s.Require().NoError(err)
	callback, err := keeper.GetCallback(ctx, reqMsg.CallbackHeight, reqMsg.ContractAddress, reqMsg.JobId)
	s.Require().NoError(err)
	senderBalance := s.chain.GetBalance(sdk.MustAccAddressFromBech32(callback.ReservedBy))

	testCases := []struct {
		testCase    string
		input       func() *types.MsgCancelCallback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgCancelCallback {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: callback does not exist",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           2,
					CallbackHeight:  130,
					Sender:          contractAdminAcc.Address.String(),
				}
			},
			expectError: true,
			errorType:   types.ErrCallbackNotFound,
		},
		{
			testCase: "FAIL: sender is not authorized to cancel callback",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  130,
					Sender:          s.chain.GetAccount(3).Address.String(),
				}
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "OK: successfully cancel callback",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  130,
					Sender:          contractAdminAcc.Address.String(),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := msgServer.CancelCallback(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorIs(err, tc.errorType)
			} else {
				s.Require().NoError(err)
				// Ensuring the callback no longer exists
				exists, err := keeper.ExistsCallback(ctx, req.CallbackHeight, req.ContractAddress, req.JobId)
				s.Require().NoError(err)
				s.Require().False(exists)
				// Ensuring the refund amount matches expected amount
				refundAmount := callback.FeeSplit.TransactionFees.Add(*callback.FeeSplit.SurplusFees)
				s.Require().Equal(refundAmount, res.Refund)
				// Ensuring the sender's balance has been updated
				senderBalance = senderBalance.Add(refundAmount)
				s.Require().Equal(senderBalance, s.chain.GetBalance(sdk.MustAccAddressFromBech32(req.Sender)))
			}
		})
	}
}
