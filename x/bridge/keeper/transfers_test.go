package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestInboundTransfer() {
	accounts := apptesting.CreateRandomAccounts(10)

	testCases := []struct {
		name           string                   // test name
		msg            types.MsgInboundTransfer // msg to send; NB! sender is set from the msgSenders slice
		msgSenders     []string                 // sender of the msg; the msg is executed len(msgSenders) times
		moduleSigners  []string                 // authorized module signers
		assets         []types.Asset            // module assets
		votesNeeded    uint64                   // votes needed module param
		expectedErrors []error                  // expected errors, one per each msgSenders value
		finalized      bool                     // expected finalization flag
	}{
		{
			name: "transfer finalized: one signer, one vote",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{nil},
			finalized:      true,
		},
		{
			name: "transfer finalized: two signers, one vote needed",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
				accounts[1].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
				accounts[1].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{nil, nil},
			finalized:      true,
		},
		{
			name: "transfer finalized: one signer, two votes needed",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    2,
			expectedErrors: []error{nil},
			finalized:      false,
		},
		{
			name: "transfer not finalized: one signer, two votes needed",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    2,
			expectedErrors: []error{nil},
			finalized:      false,
		},
		{
			name: "error: invalid message",
			msg: types.MsgInboundTransfer{
				ExternalId:     "", // invalid external id
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{sdkerrors.ErrInvalidRequest},
			finalized:      false,
		},
		{
			name: "error: sender is not part of the signer set",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(), // sender is not part of the signer set
			},
			moduleSigners: []string{
				accounts[4].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{sdkerrors.ErrorInvalidSigner},
			finalized:      false,
		},
		{
			name: "error: unknown asset id",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId: types.AssetID{ // unknown asset id
					SourceChain: "aaa",
					Denom:       "bbb",
				},
				Amount: math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{types.ErrInvalidAssetID},
			finalized:      false,
		},
		{
			name: "error: double voting",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(), // two messages with by one sender
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    2,
			expectedErrors: []error{nil, types.ErrCantFinalizeTransfer}, // err on the second vote
			finalized:      false,
		},
		{
			name: "error: double voting",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSenders slice
				DestAddr:       accounts[9].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSenders: []string{
				accounts[0].String(),
				accounts[0].String(),
			},
			moduleSigners: []string{
				accounts[0].String(),
			},
			assets:         []types.Asset{asset1},
			votesNeeded:    2,
			expectedErrors: []error{nil, types.ErrCantFinalizeTransfer}, // err on the second vote
			finalized:      false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Prepare the app
			s.AppendNewSigners(tc.moduleSigners...)
			s.AppendNewAssets(tc.assets...)
			s.EnableAssets(types.Map(tc.assets, func(v types.Asset) types.AssetID { return v.Id })...)
			s.SetVotesNeeded(tc.votesNeeded)

			s.Require().Equal(len(tc.expectedErrors), len(tc.msgSenders),
				"The number of expected errors must be equal to the number of msg signers!")

			// Send len(msgSenders) messages
			for i := range tc.msgSenders {
				_, err := s.msgServer.InboundTransfer(s.Ctx, &types.MsgInboundTransfer{
					ExternalId:     tc.msg.ExternalId,
					ExternalHeight: tc.msg.ExternalHeight,
					Sender:         tc.msgSenders[i], // NB! signer from the msgSenders list, not from the msg
					DestAddr:       tc.msg.DestAddr,
					AssetId:        tc.msg.AssetId,
					Amount:         tc.msg.Amount,
				})
				s.Require().ErrorIs(err, tc.expectedErrors[i])
			}

			// Prepare results
			finalized := s.App.BridgeKeeper.IsTransferFinalized(s.Ctx, tc.msg.ExternalId)
			height := s.GetLastTransferHeight(tc.msg.AssetId)
			destBalance := s.App.BankKeeper.GetBalance(
				s.Ctx,
				s.GetAddrFromBech32(tc.msg.DestAddr),
				s.GetTFDenom(tc.msg.AssetId),
			)
			mintedCoins := sdk.NewCoin(s.GetTFDenom(tc.msg.AssetId), tc.msg.Amount)

			// Check results
			switch tc.finalized {
			case true:
				s.Require().True(finalized)
				s.Require().Equal(height, tc.msg.ExternalHeight)
				s.Require().True(destBalance.IsEqual(mintedCoins))
			case false:
				s.Require().False(finalized)
				s.Require().NotEqual(height, tc.msg.ExternalHeight)
				s.Require().True(destBalance.IsZero())
			}
		})
	}
}

func (s *KeeperTestSuite) TestOutboundTransfer() {
	accounts := apptesting.CreateRandomAccounts(10)

	testCases := []struct {
		name           string
		msg            types.MsgOutboundTransfer
		assets         []types.Asset // module assets
		initialBalance math.Int      // initial sender's balance of coins to transfer
		expectedError  error
	}{
		{
			name: "success",
			msg: types.MsgOutboundTransfer{
				Sender:   accounts[0].String(),
				DestAddr: accounts[9].String(),
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			assets:         []types.Asset{asset1},
			initialBalance: math.NewInt(120),
			expectedError:  nil,
		},
		{
			name: "error: invalid message",
			msg: types.MsgOutboundTransfer{
				Sender:   accounts[0].String(),
				DestAddr: "wqerqwer", // invalid addr
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			assets:         []types.Asset{asset1},
			initialBalance: math.NewInt(120),
			expectedError:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "error: insufficient initial balance",
			msg: types.MsgOutboundTransfer{
				Sender:   accounts[0].String(),
				DestAddr: accounts[9].String(),
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			assets:         []types.Asset{asset1},
			initialBalance: math.NewInt(80), // initial balance is less than the transfer amount
			expectedError:  types.ErrTokenfactory,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Prepare the app
			s.AppendNewAssets(tc.assets...)
			s.EnableAssets(types.Map(tc.assets, func(v types.Asset) types.AssetID { return v.Id })...)
			initialCoins := sdk.NewCoin(s.GetTFDenom(tc.msg.AssetId), tc.initialBalance)
			_, err := s.RunMsg(&tokenfactorytypes.MsgMint{
				Sender:        s.GetModuleAddress(),
				Amount:        initialCoins,
				MintToAddress: tc.msg.Sender,
			})
			s.Require().NoError(err)

			// Send a message
			_, err = s.msgServer.OutboundTransfer(s.Ctx, &types.MsgOutboundTransfer{
				Sender:   tc.msg.Sender,
				DestAddr: tc.msg.DestAddr,
				AssetId:  tc.msg.AssetId,
				Amount:   tc.msg.Amount,
			})
			s.Require().ErrorIs(err, tc.expectedError)

			// Prepare results
			destBalance := s.App.BankKeeper.GetBalance(
				s.Ctx,
				s.GetAddrFromBech32(tc.msg.Sender),
				s.GetTFDenom(tc.msg.AssetId),
			)
			burnedCoins := sdk.NewCoin(s.GetTFDenom(tc.msg.AssetId), tc.msg.Amount)

			// Check results
			switch {
			case tc.expectedError != nil:
				// Balance hasn't changed
				s.Require().True(initialCoins.IsEqual(destBalance))
			case tc.expectedError == nil:
				// Balance has changed
				s.Require().True(initialCoins.Sub(burnedCoins).IsEqual(destBalance))
			}
		})
	}
}
