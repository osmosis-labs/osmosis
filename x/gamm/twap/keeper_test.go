package twap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

var defaultUniV2Coins = sdk.NewCoins(sdk.NewInt64Coin("token/B", 100), sdk.NewInt64Coin("token/A", 100))

type TestSuite struct {
	apptesting.KeeperTestHelper
	twapkeeper *twap.Keeper
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
	suite.Setup()
	suite.twapkeeper = suite.App.TwapKeeper
}

func newEmptyPriceRecord(poolId uint64, t time.Time, asset0 string, asset1 string) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,
	}
}
