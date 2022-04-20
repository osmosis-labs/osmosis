package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/claim/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	airdropStartTime := suite.Ctx.BlockHeader().Time
	suite.App.ClaimKeeper.CreateModuleAccount(suite.Ctx, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)))

	err := suite.App.ClaimKeeper.SetParams(suite.Ctx, types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
		ClaimDenom:         sdk.DefaultBondDenom,
	})
	if err != nil {
		panic(err)
	}

	suite.Ctx = suite.Ctx.WithBlockTime(airdropStartTime)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
