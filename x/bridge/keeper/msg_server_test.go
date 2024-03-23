package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (s *KeeperTestSuite) TestInboundTransfer() {
	testCases := []struct {
		name           string                   // test name
		msg            types.MsgInboundTransfer // msg to send; NB! sender is set from the msgSigners slice
		msgSigners     []string                 // signers of the msg; the msg is executed len(msgSigners) times
		moduleSigners  []string                 // authorized module signers
		assets         []types.Asset            // module assets
		votesNeeded    uint64                   // votes needed module param
		expectedErrors []error                  // expected errors, one per each msgSigners value
		finalized      bool                     // expected finalization flag
	}{
		{
			name: "success",
			msg: types.MsgInboundTransfer{
				ExternalId:     "unique_external_id",
				ExternalHeight: 100,
				Sender:         "", // sender is set from the msgSigners slice
				DestAddr:       s.TestAccs[1].String(),
				AssetId:        assetID1,
				Amount:         math.NewInt(100),
			},
			msgSigners:     []string{s.TestAccs[0].String()},
			moduleSigners:  []string{s.TestAccs[0].String()},
			assets:         []types.Asset{asset1},
			votesNeeded:    1,
			expectedErrors: []error{nil},
			finalized:      true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			s.AppendNewSigners(tc.moduleSigners...)
			s.AppendNewAssets(tc.assets...)
			s.EnableAssets(types.Map(tc.assets, func(v types.Asset) types.AssetID { return v.Id })...)
			s.SetVotesNeeded(tc.votesNeeded)

			s.Require().Equal(len(tc.expectedErrors), len(tc.msgSigners),
				"The number of expected errors must be equal to the number of msg signers!")

			for i := range tc.msgSigners {
				_, err := s.msgServer.InboundTransfer(s.Ctx, &types.MsgInboundTransfer{
					ExternalId:     tc.msg.ExternalId,
					ExternalHeight: tc.msg.ExternalHeight,
					Sender:         tc.msgSigners[i], // NB! signer from the msgSigners list, not from the msg
					DestAddr:       tc.msg.DestAddr,
					AssetId:        tc.msg.AssetId,
					Amount:         tc.msg.Amount,
				})
				s.Require().ErrorIs(err, tc.expectedErrors[i])
			}

			finalized := s.App.BridgeKeeper.IsTransferFinalized(s.Ctx, tc.msg.ExternalId)
			height := s.GetLastTransferHeight(tc.msg.AssetId)
			balanceMinted := s.App.BankKeeper.HasBalance(
				s.Ctx,
				s.GetAddrFromBech32(tc.msg.DestAddr),
				sdk.NewCoin(s.GetTFDenom(tc.msg.AssetId), tc.msg.Amount),
			)

			switch tc.finalized {
			case true:
				s.Require().True(finalized)
				s.Require().Equal(height, tc.msg.ExternalHeight)
				s.Require().True(balanceMinted)
			case false:
				s.Require().False(finalized)
				s.Require().NotEqual(height, tc.msg.ExternalHeight)
				s.Require().False(balanceMinted)
			}
		})
	}
}
