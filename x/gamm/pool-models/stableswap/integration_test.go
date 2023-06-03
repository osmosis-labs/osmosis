// This file contains integration tests, using "true" messages.
// We expect tests for:
// * MsgCreatePool creating correct pool as expected
// * MsgStableSwapAdjustScalingFactors works as expected
package stableswap_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
	suite.Setup()
}

func (s *TestSuite) TestSetScalingFactors() {
	s.SetupTest()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())
	nextPoolId := s.App.GAMMKeeper.GetNextPoolId(s.Ctx)
	defaultCreatePoolMsg := *baseCreatePoolMsgGen(addr1)
	defaultCreatePoolMsg.ScalingFactorController = defaultCreatePoolMsg.Sender
	defaultAdjustSFMsg := stableswap.NewMsgStableSwapAdjustScalingFactors(defaultCreatePoolMsg.Sender, nextPoolId, []uint64{1, 1})

	tests := map[string]struct {
		createMsg  stableswap.MsgCreateStableswapPool
		setMsg     stableswap.MsgStableSwapAdjustScalingFactors
		expectPass bool
	}{
		"valid_msg": {defaultCreatePoolMsg, defaultAdjustSFMsg, true},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			sender := tc.createMsg.GetSigners()[0]
			s.FundAcc(sender, s.App.GAMMKeeper.GetParams(s.Ctx).PoolCreationFee)
			s.FundAcc(sender, tc.createMsg.InitialPoolLiquidity.Sort())
			_, err := s.RunMsg(&tc.createMsg)
			s.Require().NoError(err)
			_, err = s.RunMsg(&tc.setMsg)
			s.Require().NoError(err)
		})
	}
}
