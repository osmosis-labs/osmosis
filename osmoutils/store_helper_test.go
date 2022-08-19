package osmoutils_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
	store sdk.KVStore
}

const (
	keyA               = "a"
	keyB               = "b"
	keyC               = "c"
	mockStopValue      = "stop"
	afterMockStopValue = mockStopValue + keyA
	basePrefix         = "base"
	prefixOne          = "one"
	prefixTwo          = "two"
)

func TestOsmoUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupStoreWithBasePrefix() {
	_, ms := s.CreateTestContextWithMultiStore()
	prefix := sdk.NewKVStoreKey(basePrefix)
	ms.MountStoreWithDB(prefix, sdk.StoreTypeIAVL, nil)
	err := ms.LoadLatestVersion()
	s.Require().NoError(err)
	s.store = ms.GetKVStore(prefix)
}

func mockParseValue(b []byte) (string, error) {
	return string(b), nil
}

func mockParseValueWithError(b []byte) (string, error) {
	return "", errors.New("mock error")
}

func mockStop(b []byte) bool {
	return string(b) == fmt.Sprintf("%s%s", prefixOne, mockStopValue)
}

func (s *TestSuite) TestGatherValuesFromIteratorWithStop() {

	testcases := map[string]struct {
		// if prefix is set, startValue and endValue are ignored.
		// we either create an iterator prefix or a range iterator.
		prefix     string
		startValue string
		endValue   string
		preSetKeys []string
		isReverse  bool

		expectedValues []string
		expectedErr    bool
	}{
		"prefix iterator, no stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},

			prefix: prefixOne,

			expectedValues: []string{"0", "1", "2"},
		},
		"prefix iterator, with stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + mockStopValue + keyA},
			prefix:     prefixOne,

			expectedValues: []string{"0"},
		},
		"prefix iterator, with stop, different insertion order": {
			// keyB is lexicographically before mockStopValue so it is returned, but before c
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + keyB, prefixOne + mockStopValue + keyA},
			prefix:     prefixOne,

			expectedValues: []string{"0", "2"},
		},
		"range iterator, no end, no stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},

			startValue: prefixOne + keyB,

			expectedValues: []string{"1", "2"},
		},
		"range iterator, no start, no stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},

			endValue: prefixOne + keyB,

			expectedValues: []string{"0"},
		},
		"range iterator, no start no end, no stop": {
			preSetKeys:     []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},
			expectedValues: []string{"0", "1", "2"},
		},
		"range iterator, with stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + afterMockStopValue},

			expectedValues: []string{"0"},
		},
		"range iterator, reverse": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},
			isReverse:  true,

			expectedValues: []string{"2", "1", "0"},
		},
		"range iterator, other prefix is excluded with end value": {
			preSetKeys: []string{prefixOne + keyA, prefixTwo + keyA, prefixOne + keyB, prefixTwo + keyB, prefixOne + keyC, prefixTwo + keyC},
			startValue: prefixOne + keyB,
			endValue:   prefixOne + "d",
			isReverse:  true,

			expectedValues: []string{"4", "2"},
		},
		"parse with error": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC},

			prefix: prefixOne,

			expectedErr: true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupStoreWithBasePrefix()

			var iterator sdk.Iterator

			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			if tc.prefix != "" {
				iterator = sdk.KVStorePrefixIterator(s.store, []byte(tc.prefix))
			} else {
				var startValue, endValue []byte
				if tc.startValue != "" {
					startValue = []byte(tc.startValue)
				}
				if tc.endValue != "" {
					endValue = []byte(tc.endValue)
				}

				if tc.isReverse {
					iterator = s.store.ReverseIterator(startValue, endValue)
				} else {
					iterator = s.store.Iterator(startValue, endValue)
				}
				defer iterator.Close()
			}

			mockParseValueFn := mockParseValue
			if tc.expectedErr {
				mockParseValueFn = mockParseValueWithError
			}

			actualValues, err := osmoutils.GatherValuesFromIteratorWithStop(iterator, mockParseValueFn, mockStop)

			if tc.expectedErr {
				s.Require().Error(err)
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)

			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}
