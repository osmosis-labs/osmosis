package writelistener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/indexer/domain"
	indexerdomain "github.com/osmosis-labs/osmosis/v28/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/indexer/domain/mocks"
	"github.com/osmosis-labs/osmosis/v28/ingest/indexer/service/writelistener"
)

type WriteListenerTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

var (
	someValue = []byte("someValue")
)

func TestWriteListenerTestSuite(t *testing.T) {
	suite.Run(t, new(WriteListenerTestSuite))
}

// Validates the OnWrite method of the bank write listener.
// Configures the token supply publisher mock to either return an expected forcer error
// or validates that the mock's relevant publish methods were called with the desired arguments.
func (s *WriteListenerTestSuite) TestWriteListener_Bank() {

	const (
		defaultDenom = "noErrorDenom"
	)

	var (
		oneInt         = osmomath.NewInt(1)
		oneIntBytes, _ = oneInt.Marshal()

		miscKVKey = []byte("miscKVKey")

		defaultError = errors.New("defaultError")
	)

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		hasColdStarted bool

		forceTokenSupplyError       error
		ForceTokenSupplyOffsetError error

		expectedCalledPublishTokenSupply       indexerdomain.TokenSupply
		expectedCalledPublishTokenSupplyOffset indexerdomain.TokenSupplyOffset

		expectedError error
	}{
		{
			name:  "published supply key",
			key:   append(banktypes.SupplyKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			hasColdStarted: true,

			expectedCalledPublishTokenSupply: indexerdomain.TokenSupply{
				Denom:  defaultDenom,
				Supply: oneInt,
			},
		},
		{
			name:  "published supply offset key",
			key:   append(banktypes.SupplyOffsetKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			hasColdStarted: true,

			expectedCalledPublishTokenSupplyOffset: indexerdomain.TokenSupplyOffset{
				Denom:        defaultDenom,
				SupplyOffset: oneInt,
			},
		},
		{
			name:  "did non publish supply key before cold start",
			key:   append(banktypes.SupplyKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			expectedError: domain.ErrDidNotIngestAllData,
		},
		{
			name:  "did not publish supply offset key before cold start",
			key:   append(banktypes.SupplyOffsetKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			expectedError: domain.ErrDidNotIngestAllData,
		},
		{
			name:  "published nothing due to misc key",
			key:   append(miscKVKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			hasColdStarted: true,
		},

		{
			name:  "forced supply error by client mock",
			key:   append(banktypes.SupplyKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			hasColdStarted: true,

			forceTokenSupplyError: defaultError,

			expectedError: defaultError,
		},
		{
			name:  "forced supply offset error by client mock",
			key:   append(banktypes.SupplyOffsetKey, []byte(defaultDenom)...),
			value: oneIntBytes,

			hasColdStarted: true,

			ForceTokenSupplyOffsetError: defaultError,

			expectedError: defaultError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			s.Setup()

			// Initialize cold start manager
			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			// Mark initial data ingested if the test case has cold started.
			if tc.hasColdStarted {
				blockProcessStrategyManager.MarkInitialDataIngested()
			}

			// Initialize token supply publisher mock
			tokenSupplyPublisherMock := &mocks.PublisherMock{
				ForceTokenSupplyError:       tc.forceTokenSupplyError,
				ForceTokenSupplyOffsetError: tc.ForceTokenSupplyOffsetError,
			}

			// Initialize bank write listener
			bankWriteListener := writelistener.NewBank(context.TODO(), tokenSupplyPublisherMock, blockProcessStrategyManager)

			bankKVStore := s.App.GetKey(banktypes.ModuleName)

			// System under test
			err := bankWriteListener.OnWrite(bankKVStore, tc.key, tc.value, tc.isDelete)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}

			// Validation.
			s.Require().NoError(err)

			s.Require().Equal(tc.expectedCalledPublishTokenSupply, tokenSupplyPublisherMock.CalledWithTokenSupply)
			s.Require().Equal(tc.expectedCalledPublishTokenSupplyOffset, tokenSupplyPublisherMock.CalledWithTokenSupplyOffset)
		})
	}
}
