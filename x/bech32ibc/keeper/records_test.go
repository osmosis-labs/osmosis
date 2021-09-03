package keeper_test

import (
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
)

func (suite *KeeperTestSuite) TestNativeHrpLifeCycle() {
	suite.SetupTest()

	// check genesis native hrp
	nativeHrp, err := suite.app.Bech32IBCKeeper.GetNativeHrp(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(nativeHrp, "osmo")

	// check update of native hrp correctly
	err = suite.app.Bech32IBCKeeper.SetNativeHrp(suite.ctx, "osmosis")
	suite.Require().NoError(err)

	nativeHrp, err = suite.app.Bech32IBCKeeper.GetNativeHrp(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(nativeHrp, "osmosis")

	// error for uppercase in denom
	err = suite.app.Bech32IBCKeeper.SetNativeHrp(suite.ctx, "OSMO")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestHrpIbcRecordsLifeCycle() {
	suite.SetupTest()

	// check genesis hrp ibc records
	hrpIbcRecords := suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 0)

	// check update of native hrp correctly
	suite.app.Bech32IBCKeeper.SetHrpIbcRecords(suite.ctx, []types.HrpIbcRecord{
		{
			Hrp:           "akash",
			SourceChannel: "channel-1",
		},
	})

	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 1)

	// check update of native hrp correctly
	suite.app.Bech32IBCKeeper.SetHrpIbcRecords(suite.ctx, []types.HrpIbcRecord{
		{
			Hrp:           "cosmos",
			SourceChannel: "channel-2",
		},
		{
			Hrp:           "iris",
			SourceChannel: "channel-3",
		},
	})
	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 3)

	// check update twice
	suite.app.Bech32IBCKeeper.SetHrpIbcRecords(suite.ctx, []types.HrpIbcRecord{
		{
			Hrp:           "cosmos",
			SourceChannel: "channel-5",
		},
	})
	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 3)

	// check deletion
	suite.app.Bech32IBCKeeper.SetHrpIbcRecords(suite.ctx, []types.HrpIbcRecord{
		{
			Hrp:           "cosmos",
			SourceChannel: "",
		},
	})
	hrpIbcRecords = suite.app.Bech32IBCKeeper.GetHrpIbcRecords(suite.ctx)
	suite.Require().Len(hrpIbcRecords, 2)

	// Check getting individually
	hrpIbcRecord, err := suite.app.Bech32IBCKeeper.GetHrpIbcRecord(suite.ctx, "cosmos")
	suite.Require().Error(err)

	hrpIbcRecord, err = suite.app.Bech32IBCKeeper.GetHrpIbcRecord(suite.ctx, "akash")
	suite.Require().Equal(hrpIbcRecord, types.HrpIbcRecord{
		Hrp:           "akash",
		SourceChannel: "channel-1",
	})
	sourceChan, err := suite.app.Bech32IBCKeeper.GetHrpSourceChannel(suite.ctx, "akash")
	suite.Require().NoError(err)
	suite.Require().Equal(sourceChan, "channel-1")

	suite.Require().NoError(err)
	hrpIbcRecord, err = suite.app.Bech32IBCKeeper.GetHrpIbcRecord(suite.ctx, "iris")
	suite.Require().Equal(hrpIbcRecord, types.HrpIbcRecord{
		Hrp:           "iris",
		SourceChannel: "channel-3",
	})
	suite.Require().NoError(err)

	hrpIbcRecord, err = suite.app.Bech32IBCKeeper.GetHrpIbcRecord(suite.ctx, "11")
	suite.Require().Error(err)
	_, err = suite.app.Bech32IBCKeeper.GetHrpSourceChannel(suite.ctx, "11")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestValidateHrpIbcRecord() {
	suite.SetupTest()

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

	// check invalid prefix for hrp
	err := suite.app.Bech32IBCKeeper.ValidateHrpIbcRecord(suite.ctx, types.HrpIbcRecord{
		Hrp:           "AAA",
		SourceChannel: "channel-1",
	})
	suite.Require().Error(err)

	// check native hrp
	err = suite.app.Bech32IBCKeeper.ValidateHrpIbcRecord(suite.ctx, types.HrpIbcRecord{
		Hrp:           "osmo",
		SourceChannel: "channel-1",
	})
	suite.Require().Error(err)

	// check invalid channel
	err = suite.app.Bech32IBCKeeper.ValidateHrpIbcRecord(suite.ctx, types.HrpIbcRecord{
		Hrp:           "osmo",
		SourceChannel: "channel-2",
	})
	suite.Require().Error(err)

	// check correct one
	err = suite.app.Bech32IBCKeeper.ValidateHrpIbcRecord(suite.ctx, types.HrpIbcRecord{
		Hrp:           "akash",
		SourceChannel: "channel-1",
	})
	suite.Require().NoError(err)
}
