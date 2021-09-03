package keeper_test

import (
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
)

func (suite *KeeperTestSuite) TestHandleUpdateHrpIbcChannelProposal() {
	suite.SetupTest()

	// check genesis hrp ibc records
	hrpIbcRecords := suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 0)

	// check valid channel existence correctly
	err := suite.app.Bech32IBCKeeper.HandleUpdateHrpIbcChannelProposal(suite.ctx, &types.UpdateHrpIbcChannelProposal{
		Title:         "update hrp ibc channel",
		Description:   "update hrp ibc channel",
		Hrp:           "akash",
		SourceChannel: "channel-1",
	})
	suite.Require().Error(err)

	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 0)

	// create channel and try
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, "transfer", "channel-1", channeltypes.Channel{
		State:    1,
		Ordering: 1,
		Counterparty: channeltypes.Counterparty{
			PortId:    "transfer",
			ChannelId: "channel-1",
		},
		ConnectionHops: []string{},
		Version:        "ics20",
	})
	err = suite.app.Bech32IBCKeeper.HandleUpdateHrpIbcChannelProposal(suite.ctx, &types.UpdateHrpIbcChannelProposal{
		Title:         "update hrp ibc channel",
		Description:   "update hrp ibc channel",
		Hrp:           "akash",
		SourceChannel: "channel-1",
	})
	suite.Require().NoError(err)

	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 1)
}
