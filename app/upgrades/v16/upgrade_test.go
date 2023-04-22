package v16_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v15/app/params"

	v16 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v16"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpdateTokenFactoryParams() {
	suite.SetupTest() // reset

	ctx := suite.Ctx
	tokenFactoryKeeper := suite.App.TokenFactoryKeeper

	// before migration
	params := tokenFactoryKeeper.GetParams(ctx)
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdk.NewInt(10_000_000))), params.DenomCreationFee)
	suite.Require().Equal(uint64(0), params.DenomCreationGasConsume)

	v16.UpdateTokenFactoryParams(ctx, tokenFactoryKeeper)

	// after migration
	params = tokenFactoryKeeper.GetParams(ctx)
	suite.Require().Nil(params.DenomCreationFee)
	suite.Require().Equal(v16.NewDenomCreationGasConsume, params.DenomCreationGasConsume)
}
