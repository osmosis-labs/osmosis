package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

func (suite *KeeperTestSuite) TestAutostakingManagement() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	addr3 := sdk.AccAddress([]byte("addr3---------------"))

	valAddr1 := sdk.ValAddress(addr1)
	valAddr2 := sdk.ValAddress(addr2)

	suite.SetupTest()
	err := suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr1.String(),
		AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr2.String(),
		AutostakingValidator: valAddr2.String(),
		AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})
	suite.Require().NoError(err)

	autostaking1 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr2.String())
	suite.Require().NotNil(autostaking1)
	suite.Require().Equal(*autostaking1, types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr1.String(),
		AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})

	autostaking2 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr2.String())
	suite.Require().NotNil(autostaking2)
	suite.Require().Equal(*autostaking2, types.AutoStaking{
		Address:              addr2.String(),
		AutostakingValidator: valAddr2.String(),
		AutostakingRate:      sdk.NewDecWithPrec(5, 1),
	})

	autostaking3 := suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr3.String())
	suite.Require().Nil(autostaking3)

	err = suite.app.IncentivesKeeper.SetAutostaking(suite.ctx, &types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr2.String(),
		AutostakingRate:      sdk.NewDecWithPrec(1, 1),
	})
	suite.Require().NoError(err)

	autostaking1 = suite.app.IncentivesKeeper.GetAutostakingByAddress(suite.ctx, addr2.String())
	suite.Require().NotNil(autostaking1)
	suite.Require().Equal(*autostaking1, types.AutoStaking{
		Address:              addr1.String(),
		AutostakingValidator: valAddr2.String(),
		AutostakingRate:      sdk.NewDecWithPrec(1, 1),
	})

	autostakings := suite.app.IncentivesKeeper.AllAutoStakings(suite.ctx)
	suite.Require().Len(autostakings, 2)

	autostakingIters := []types.AutoStaking{}
	suite.app.IncentivesKeeper.IterateAutoStaking(suite.ctx, func(index int64, autostaking types.AutoStaking) (stop bool) {
		autostakingIters = append(autostakingIters, autostaking)
		return false
	})
	suite.Require().Len(autostakingIters, 2)
}
