package osmoutils_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	twaptypes "github.com/osmosis-labs/osmosis/v11/x/twap/types"
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

func (s *TestSuite) TestNoStopFn_AlwaysFalse() {
	s.Require().False(osmoutils.NoStopFn([]byte(keyA)))
	s.Require().False(osmoutils.NoStopFn([]byte(keyB)))
}

// TestMustGet tests that MustGet retrieves the correct
// values from the store and panics if an error is encountered.
func (s *TestSuite) TestMustGet() {

	tests := map[string]struct {
		// keys and values to preset
		preSetKeyValues map[string]proto.Message

		// keys and values to attempt to get and validate
		expectedGetKeyValues map[string]proto.Message

		actualResultProto proto.Message

		expectPanic bool
	}{
		"basic valid test": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
				keyB: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
				keyC: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
				keyB: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
				keyC: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			actualResultProto: &sdk.DecProto{},
		},
		"attempt to get non-existent key - panic": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
				keyC: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyB: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			actualResultProto: &sdk.DecProto{},

			expectPanic: true,
		},
		"invalid proto Dec vs TwapRecord- error": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			actualResultProto: &twaptypes.TwapRecord{},

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupStoreWithBasePrefix()

			// Setup
			for key, value := range tc.preSetKeyValues {
				osmoutils.MustSet(s.store, []byte(key), value)
			}

			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				for key, expectedValue := range tc.expectedGetKeyValues {
					// System under test.
					osmoutils.MustGet(s.store, []byte(key), tc.actualResultProto)
					// Assertions.
					s.Require().Equal(expectedValue.String(), tc.actualResultProto.String())
				}
			})
		})
	}
}

// TestMustSet tests that MustSet updates the store correctly
// and panics if an error is encountered.
func (s *TestSuite) TestMustSet() {
	tests := map[string]struct {
		// keyys and values to preset
		setKey   string
		setValue proto.Message

		// keys and values to attempt to get and validate
		getKeyValues map[string]proto.Message

		actualResultProto proto.Message

		key         []byte
		result      proto.Message
		expectPanic bool
	}{
		"basic valid Dec test": {
			setKey: keyA,
			setValue: &sdk.DecProto{
				Dec: sdk.OneDec(),
			},

			actualResultProto: &sdk.DecProto{},
		},
		"basic valid TwapRecord test": {
			setKey: keyA,
			setValue: &twaptypes.TwapRecord{
				PoolId: 2,
			},

			actualResultProto: &twaptypes.TwapRecord{},
		},
		"invalid set value": {
			setKey:   keyA,
			setValue: (*sdk.DecProto)(nil),

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupStoreWithBasePrefix()

			// Setup
			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				osmoutils.MustSet(s.store, []byte(tc.setKey), tc.setValue)
			})

			if tc.expectPanic {
				return
			}

			osmoutils.MustGet(s.store, []byte(tc.setKey), tc.actualResultProto)
			s.Require().Equal(tc.setValue.String(), tc.actualResultProto.String())
		})
	}
}

// TestMustGetDec tests that MustGetDec retrieves the correct
// decimal values from the store and panics if an error is encountered.
func (s *TestSuite) TestMustGetDec() {

	tests := map[string]struct {
		// keyys and values to preset
		preSetKeyValues map[string]sdk.Dec

		// keys and values to attempt to get and validate
		getKeyValues map[string]sdk.Dec

		expectPanic bool
	}{
		"valid get": {
			preSetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: sdk.OneDec().Add(sdk.OneDec()),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},

			getKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: sdk.OneDec().Add(sdk.OneDec()),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},
		},
		"attempt to get non-existent key - panic": {
			preSetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},

			getKeyValues: map[string]sdk.Dec{
				keyB: sdk.OneDec().Add(sdk.OneDec()),
			},

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupStoreWithBasePrefix()

			// Setup
			for key, value := range tc.preSetKeyValues {
				osmoutils.MustSetDec(s.store, []byte(key), value)
			}

			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				for key, expectedValue := range tc.getKeyValues {
					// System under test.
					actualDec := osmoutils.MustGetDec(s.store, []byte(key))
					// Assertions.
					s.Require().Equal(expectedValue.String(), actualDec.String())
				}
			})
		})
	}
}

// TestMustSetDec tests that MustSetDec updates the store correctly
// with the right decimal value.
func (s *TestSuite) TestMustSetDec() {
	// Setup.
	s.SetupStoreWithBasePrefix()

	originalDecValue := sdk.OneDec()

	// System under test.
	osmoutils.MustSetDec(s.store, []byte(keyA), originalDecValue)

	// Assertions.
	retrievedDecVaue := osmoutils.MustGetDec(s.store, []byte(keyA))
	s.Require().Equal(originalDecValue.String(), retrievedDecVaue.String())
}
