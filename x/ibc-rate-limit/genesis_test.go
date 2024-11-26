package ibc_rate_limit_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
)

type GenesisTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	suite.SetupTest()
	k := suite.App.RateLimitingICS4Wrapper

	initialGenesis := types.GenesisState{
		Params: types.Params{
			ContractAddress: testAddress,
		},
	}

	k.InitGenesis(suite.Ctx, initialGenesis)

	suite.Require().Equal(testAddress, k.GetParams(suite.Ctx).ContractAddress)

	exportedGenesis := k.ExportGenesis(suite.Ctx)

	suite.Require().Equal(initialGenesis, *exportedGenesis)
}
