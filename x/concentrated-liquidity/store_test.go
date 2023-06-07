package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	cl "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	defaultTickIndex = 1
)

// TestParseIncentiveRecordFromBytes_KeySeparatorInAddress validates that parsing
// succeeds even if the address contains the key separator. This is ensured
// by base32 encoding of the key separator.
func (s *KeeperTestSuite) TestParseIncentiveRecordFromBytes_KeySeparatorInAddress() {
	s.SetupTest()

	expectedIncentiveRecord := types.IncentiveRecord{
		PoolId:               validPoolId,
		IncentiveDenom:       testDenomOne,
		IncentiveCreatorAddr: s.TestAccs[0].String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: defaultIncentiveAmount,
			EmissionRate:    testEmissionOne,
			StartTime:       defaultStartTime,
		},
		MinUptime: testUptimeOne,
	}

	validValueBz, err := proto.Marshal(&expectedIncentiveRecord.IncentiveRecordBody)
	s.Require().NoError(err)

	uptimeIndex, err := cl.FindUptimeIndex(expectedIncentiveRecord.MinUptime)
	s.Require().NoError(err)

	incentiveRecordKey := types.KeyIncentiveRecord(expectedIncentiveRecord.PoolId, uptimeIndex, expectedIncentiveRecord.IncentiveDenom, s.TestAccs[0])

	// System under test with basic valid record.
	record, err := cl.ParseFullIncentiveRecordFromBz(incentiveRecordKey, validValueBz)
	s.Require().NoError(err)

	s.Require().Equal(expectedIncentiveRecord, record)

	// System under test with address containing a key separator.
	addrStr := fmt.Sprintf("__________%s_________", types.KeySeparator)
	keySeparatorAddress := sdk.AccAddress(addrStr)

	expectedIncentiveRecord.IncentiveCreatorAddr = keySeparatorAddress.String()
	incentiveRecordKey = types.KeyIncentiveRecord(expectedIncentiveRecord.PoolId, uptimeIndex, expectedIncentiveRecord.IncentiveDenom, keySeparatorAddress)

	// System under test with address containing a key separator.
	record, err = cl.ParseFullIncentiveRecordFromBz(incentiveRecordKey, validValueBz)
	s.Require().NoError(err)

	s.Require().Equal(expectedIncentiveRecord, record)
}
