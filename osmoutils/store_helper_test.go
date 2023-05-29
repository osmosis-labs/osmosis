package osmoutils_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/noapptest"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
)

// We need to setup a test suite with account keeper
// and a custom store setup.
// unfortunately setting up account implies setting up params
type TestSuite struct {
	suite.Suite

	ctx   sdk.Context
	store sdk.KVStore

	authStoreKey  sdk.StoreKey
	accountKeeper authkeeper.AccountKeeperI
}

func (suite *TestSuite) SetupTest() {
	// For the test suite, we manually wire a custom store "customStoreKey"
	// Auth module (for module_account_test.go) which requires params module as well.
	customStoreKey := sdk.NewKVStoreKey("osmoutil_store_test")
	suite.authStoreKey = sdk.NewKVStoreKey(authtypes.StoreKey)
	// setup ctx + stores
	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	paramsTKey := sdk.NewKVStoreKey(paramstypes.TStoreKey)
	suite.ctx = noapptest.DefaultCtxWithStoreKeys(
		[]sdk.StoreKey{customStoreKey, suite.authStoreKey, paramsKey, paramsTKey})
	suite.store = suite.ctx.KVStore(customStoreKey)
	// setup params (needed for auth)
	encConfig := noapptest.MakeTestEncodingConfig(auth.AppModuleBasic{}, params.AppModuleBasic{})
	paramsKeeper := paramskeeper.NewKeeper(encConfig.Codec, encConfig.Amino, paramsKey, paramsTKey)
	paramsKeeper.Subspace(authtypes.ModuleName)

	// setup auth
	maccPerms := map[string][]string{
		"fee_collector": nil,
		"mint":          {"minter"},
	}
	authsubspace, _ := paramsKeeper.GetSubspace(authtypes.ModuleName)
	suite.accountKeeper = authkeeper.NewAccountKeeper(
		encConfig.Codec,
		suite.authStoreKey,
		authsubspace,
		authtypes.ProtoBaseAccount, maccPerms)
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

var (
	oneA                 = []string{prefixOne + keyA}
	oneAB                = []string{prefixOne + keyA, prefixOne + keyB}
	twoAB                = []string{prefixTwo + keyA, prefixTwo + keyB}
	oneABC               = []string{prefixOne + keyA, prefixOne + keyB, prefixOne + keyC}
	oneBCA               = []string{prefixOne + keyB, prefixOne + keyC, prefixOne + keyA}
	oneABtwoAB           = []string{prefixOne + keyA, prefixOne + keyB, prefixTwo + keyA, prefixTwo + keyB}
	oneBtwoAoneAtwoB     = []string{prefixOne + keyB, prefixTwo + keyA, prefixOne + keyA, prefixTwo + keyB}
	oneAtwoAoneBtwoB     = []string{prefixOne + keyA, prefixTwo + keyA, prefixOne + keyB, prefixTwo + keyB}
	onetwoABCalternating = []string{prefixOne + keyA, prefixTwo + keyA, prefixOne + keyB, prefixTwo + keyB, prefixOne + keyC, prefixTwo + keyC}
	mockError            = errors.New("mock error")
)

func TestOsmoUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func mockParseValue(b []byte) (string, error) {
	return string(b), nil
}

func mockParseValueWithError(b []byte) (string, error) {
	return "", mockError
}

func mockStop(b []byte) bool {
	return string(b) == fmt.Sprintf("%s%s", prefixOne, mockStopValue)
}

func mockParseWithKey(key []byte, value []byte) (string, error) {
	return string(key) + string(value), nil
}

func mockParseWithKeyError(key []byte, value []byte) (string, error) {
	return "", mockError
}

func (s *TestSuite) TestGatherAllKeysFromStore() {
	testcases := map[string]struct {
		preSetKeys     []string
		expectedValues []string
	}{
		"multiple keys in lexicographic order": {
			preSetKeys:     oneABC,
			expectedValues: oneABC,
		},
		"multiple keys out of lexicographic order": {
			preSetKeys: oneBCA,
			// we expect output to be in ascending lexicographic order
			expectedValues: oneABC,
		},
		"no keys": {
			preSetKeys:     []string{},
			expectedValues: []string{},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues := osmoutils.GatherAllKeysFromStore(s.store)

			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGatherValuesFromStore() {
	testcases := map[string]struct {
		preSetKeys []string
		keyStart   []byte
		keyEnd     []byte
		parseFn    func(b []byte) (string, error)

		expectedErr    error
		expectedValues []string
	}{
		"common prefix, exclude end": {
			preSetKeys: oneAB,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixOne + keyB),
			parseFn:  mockParseValue,

			expectedValues: []string{"0"},
		},
		"common prefix, include end": {
			preSetKeys: oneAB,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixOne + keyC),
			parseFn:  mockParseValue,

			expectedValues: []string{"0", "1"},
		},
		"different prefix, inserted in lexicographic order": {
			preSetKeys: oneABtwoAB,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixTwo + keyA),
			parseFn:  mockParseValue,

			expectedValues: []string{"0", "1"},
		},
		"different prefix, inserted out of lexicographic order": {
			preSetKeys: oneAtwoAoneBtwoB,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixTwo + keyA),
			parseFn:  mockParseValue,

			// should get all prefixOne values as keys are stored in ascending lexicographic order
			expectedValues: []string{"0", "2"},
		},
		"start key and end key same": {
			preSetKeys: oneA,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixOne + keyA),
			parseFn:  mockParseValue,

			expectedValues: []string{},
		},
		"start key after end key": {
			preSetKeys: oneABC,

			keyStart: []byte(prefixOne + keyB),
			keyEnd:   []byte(prefixOne + keyA),
			parseFn:  mockParseValue,

			expectedValues: []string{},
		},
		"get all values": {
			preSetKeys: oneABC,

			keyStart: nil,
			keyEnd:   nil,
			parseFn:  mockParseValue,

			expectedValues: []string{"0", "1", "2"},
		},
		"get all values after start key": {
			// SDK iterator is broken for nil end byte, and non-nil start byte
			// https://github.com/cosmos/cosmos-sdk/issues/12661
			// so we use []byte{0xff}
			preSetKeys: oneABC,

			keyStart: []byte(prefixOne + keyB),
			keyEnd:   []byte{0xff},
			parseFn:  mockParseValue,

			expectedValues: []string{"1", "2"},
		},
		"parse with error": {
			preSetKeys: oneABC,

			keyStart: []byte(prefixOne + keyA),
			keyEnd:   []byte(prefixOne + keyC),
			parseFn:  mockParseValueWithError,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GatherValuesFromStore(s.store, tc.keyStart, tc.keyEnd, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGatherValuesFromStorePrefix() {
	testcases := map[string]struct {
		prefix     []byte
		preSetKeys []string
		parseFn    func(b []byte) (string, error)

		expectedErr    error
		expectedValues []string
	}{
		"common prefix": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),

			parseFn: mockParseValue,

			expectedValues: []string{"0", "1", "2"},
		},
		"different prefixes in order, prefix one requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			expectedValues: []string{"0", "1"},
		},
		"different prefixes in order, prefix two requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseValue,

			expectedValues: []string{"2", "3"},
		},
		"different prefixes out of order, prefix one requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			// we expect the prefixOne values in ascending lexicographic order
			expectedValues: []string{"2", "0"},
		},
		"different prefixes out of order, prefix two requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseValue,

			expectedValues: []string{"1", "3"},
		},
		"prefix doesn't exist, no keys": {
			preSetKeys: []string{},
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			expectedValues: []string{},
		},
		"prefix doesn't exist, only keys with another prefix": {
			preSetKeys: twoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			expectedValues: []string{},
		},
		"parse with error": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValueWithError,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GatherValuesFromStorePrefix(s.store, tc.prefix, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGatherValuesFromStorePrefixWithKeyParser() {
	testcases := map[string]struct {
		prefix     []byte
		preSetKeys []string
		parseFn    func(key []byte, value []byte) (string, error)

		expectedErr    error
		expectedValues []string
	}{
		"common prefix": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),

			parseFn: mockParseWithKey,

			expectedValues: []string{oneABC[0] + "0", oneABC[1] + "1", oneABC[2] + "2"},
		},
		"different prefixes in order, prefix one requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseWithKey,

			expectedValues: []string{oneABtwoAB[0] + "0", oneABtwoAB[1] + "1"},
		},
		"different prefixes in order, prefix two requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseWithKey,

			expectedValues: []string{oneABtwoAB[2] + "2", oneABtwoAB[3] + "3"},
		},
		"different prefixes out of order, prefix one requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseWithKey,

			// we expect the prefixOne values in ascending lexicographic order
			expectedValues: []string{oneBtwoAoneAtwoB[2] + "2", oneBtwoAoneAtwoB[0] + "0"},
		},
		"different prefixes out of order, prefix two requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseWithKey,

			expectedValues: []string{oneBtwoAoneAtwoB[1] + "1", oneBtwoAoneAtwoB[3] + "3"},
		},
		"prefix doesn't exist, no keys": {
			preSetKeys: []string{},
			prefix:     []byte(prefixOne),
			parseFn:    mockParseWithKey,

			expectedValues: []string{},
		},
		"prefix doesn't exist, only keys with another prefix": {
			preSetKeys: twoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseWithKey,

			expectedValues: []string{},
		},
		"parse with error": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseWithKeyError,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GatherValuesFromStorePrefixWithKeyParser(s.store, tc.prefix, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGetFirstValueAfterPrefixInclusive() {
	testcases := map[string]struct {
		prefix     []byte
		preSetKeys []string
		parseFn    func(b []byte) (string, error)

		expectedErr    error
		expectedValues string
	}{
		"common prefix": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),

			parseFn: mockParseValue,

			expectedValues: "0",
		},
		"different prefixes in order, prefix one requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			expectedValues: "0",
		},
		"different prefixes in order, prefix two requested": {
			preSetKeys: oneABtwoAB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseValue,

			expectedValues: "2",
		},
		"different prefixes out of order, prefix one requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			// we expect the prefixOne values in ascending lexicographic order
			expectedValues: "2",
		},
		"different prefixes out of order, prefix two requested": {
			preSetKeys: oneBtwoAoneAtwoB,
			prefix:     []byte(prefixTwo),
			parseFn:    mockParseValue,

			expectedValues: "1",
		},
		"prefix doesn't exist, start key lexicographically before existing keys": {
			preSetKeys: twoAB,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			// we expect the first value after the prefix, which is the value associated with the first valid key
			expectedValues: "0",
		},

		// error catching
		"prefix doesn't exist, no keys": {
			preSetKeys: []string{},
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValue,

			expectedErr:    errors.New("No values in range"),
			expectedValues: "",
		},
		"prefix doesn't exist, start key lexicographically after existing keys": {
			preSetKeys: twoAB,
			prefix:     []byte{0xff},
			parseFn:    mockParseValue,

			expectedErr:    errors.New("No values in range"),
			expectedValues: "",
		},
		"parse with error": {
			preSetKeys: oneABC,
			prefix:     []byte(prefixOne),
			parseFn:    mockParseValueWithError,

			expectedErr:    mockError,
			expectedValues: "",
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GetFirstValueAfterPrefixInclusive(s.store, tc.prefix, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Equal(tc.expectedValues, actualValues)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGatherValuesFromIterator() {
	testcases := map[string]struct {
		// if prefix is set, startValue and endValue are ignored.
		// we either create an iterator prefix or a range iterator.
		prefix     string
		startValue string
		endValue   string
		preSetKeys []string
		isReverse  bool

		expectedValues []string
		expectedErr    error
	}{
		"prefix iterator, no stop": {
			preSetKeys: oneABC,

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
			preSetKeys: oneABC,

			startValue: prefixOne + keyB,

			expectedValues: []string{"1", "2"},
		},
		"range iterator, no start, no stop": {
			preSetKeys: oneABC,

			endValue: prefixOne + keyB,

			expectedValues: []string{"0"},
		},
		"range iterator, no start no end, no stop": {
			preSetKeys:     oneABC,
			expectedValues: []string{"0", "1", "2"},
		},
		"range iterator, with stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + afterMockStopValue},

			expectedValues: []string{"0"},
		},
		"range iterator, reverse": {
			preSetKeys: oneABC,
			isReverse:  true,

			expectedValues: []string{"2", "1", "0"},
		},
		"range iterator, other prefix is excluded with end value": {
			preSetKeys: onetwoABCalternating,
			startValue: prefixOne + keyB,
			endValue:   prefixOne + "d",
			isReverse:  true,

			expectedValues: []string{"4", "2"},
		},
		"parse with error": {
			preSetKeys: oneABC,

			prefix: prefixOne,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
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
			if tc.expectedErr != nil {
				mockParseValueFn = mockParseValueWithError
			}

			actualValues, err := osmoutils.GatherValuesFromIterator(iterator, mockParseValueFn, mockStop)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)

			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGetIterValuesWithStop() {
	testcases := map[string]struct {
		preSetKeys []string
		keyStart   []byte
		keyEnd     []byte
		parseFn    func(b []byte) (string, error)
		stopFn     func(b []byte) bool
		isReverse  bool

		expectedValues []string
		expectedErr    error
	}{
		"prefix iterator, no stop but exclusive key end": {
			preSetKeys: oneABC,
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte(prefixOne + keyC),
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  false,

			expectedValues: []string{"0", "1"},
		},
		"prefix iterator, no stop and inclusive key end": {
			preSetKeys: oneAB,
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte(prefixOne + keyC),
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  false,

			expectedValues: []string{"0", "1"},
		},
		"prefix iterator, with stop before end key": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + mockStopValue + keyA},
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte(prefixOne + keyC),
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  false,

			expectedValues: []string{"0"},
		},
		"prefix iterator, with end key before stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + keyB, prefixOne + mockStopValue},
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte(prefixOne + keyB),
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  false,

			expectedValues: []string{"0"},
		},
		"prefix iterator, with stop, different insertion order": {
			// keyB is lexicographically before mockStopValue so we expect it to be returned before we hit the stopper
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + keyB, prefixOne + mockStopValue + keyA},
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte{0xff},
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  false,

			expectedValues: []string{"0", "2"},
		},
		"prefix iterator with stop, different insertion order, and reversed iterator": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + keyB, prefixOne + mockStopValue + keyA},
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte{0xff},
			parseFn:    mockParseValue,
			stopFn:     mockStop,
			isReverse:  true,

			// only the last value in our preSetKeys should be on the other end of the stopper
			expectedValues: []string{"3"},
		},
		"parse with error": {
			preSetKeys: oneABC,
			keyStart:   []byte(prefixOne + keyA),
			keyEnd:     []byte{0xff},
			parseFn:    mockParseValueWithError,
			stopFn:     mockStop,
			isReverse:  false,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GetIterValuesWithStop(s.store, tc.keyStart, tc.keyEnd, tc.isReverse, tc.stopFn, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().Nil(actualValues)
				return
			}

			s.Require().NoError(err)

			s.Require().Equal(tc.expectedValues, actualValues)
		})
	}
}

func (s *TestSuite) TestGetValuesUntilDerivedStop() {
	testcases := map[string]struct {
		preSetKeys []string
		keyStart   []byte
		parseFn    func(b []byte) (string, error)
		stopFn     func(b []byte) bool

		expectedValues []string
		expectedErr    error
	}{
		"prefix iterator, no stop": {
			preSetKeys: oneABC,
			keyStart:   []byte(prefixOne + keyA),
			parseFn:    mockParseValue,
			stopFn:     mockStop,

			expectedValues: []string{"0", "1", "2"},
		},
		"prefix iterator, with stop": {
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + mockStopValue + keyA},
			keyStart:   []byte(prefixOne + keyA),
			parseFn:    mockParseValue,
			stopFn:     mockStop,

			expectedValues: []string{"0"},
		},
		"prefix iterator, with stop & different insertion order": {
			// keyB is lexicographically before mockStopValue so we expect it to be returned before we hit the stopper
			preSetKeys: []string{prefixOne + keyA, prefixOne + mockStopValue, prefixOne + keyB, prefixOne + mockStopValue + keyA},
			keyStart:   []byte(prefixOne + keyA),
			parseFn:    mockParseValue,
			stopFn:     mockStop,

			expectedValues: []string{"0", "2"},
		},
		"parse with error": {
			preSetKeys: oneABC,
			keyStart:   []byte(prefixOne + keyA),
			parseFn:    mockParseValueWithError,
			stopFn:     mockStop,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			actualValues, err := osmoutils.GetValuesUntilDerivedStop(s.store, tc.keyStart, tc.stopFn, tc.parseFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
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
		"invalid proto Dec vs AuthParams- error": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			actualResultProto: &authtypes.Params{},

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
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

// TestGet tests that Get returns a boolean indicating
// whether value exists for the given key and error
func (s *TestSuite) TestGet() {
	tests := map[string]struct {
		// keys and values to preset
		preSetKeyValues map[string]proto.Message

		// keys and values to attempt to get and validate
		expectedGetKeyValues map[string]proto.Message

		actualResultProto proto.Message

		expectFound bool

		expectErr bool
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

			expectFound: true,
		},
		"attempt to get non-existent key - not found & no err return": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
				keyC: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyB: &sdk.DecProto{Dec: sdk.OneDec().Add(sdk.OneDec())},
			},

			actualResultProto: &sdk.DecProto{},

			expectFound: false,

			expectErr: false,
		},
		"invalid proto Dec vs AuthParams - found but Unmarshal err": {
			preSetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			expectedGetKeyValues: map[string]proto.Message{
				keyA: &sdk.DecProto{Dec: sdk.OneDec()},
			},

			actualResultProto: &authtypes.Params{},

			expectFound: true,

			expectErr: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			// Setup
			for key, value := range tc.preSetKeyValues {
				osmoutils.MustSet(s.store, []byte(key), value)
			}

			for key, expectedValue := range tc.expectedGetKeyValues {
				// System under test.
				found, err := osmoutils.Get(s.store, []byte(key), tc.actualResultProto)
				// Assertions.
				s.Require().Equal(found, tc.expectFound)
				if tc.expectErr {
					s.Require().Error(err)
				}
				// make sure found by key & Unmarshal successfully
				if !tc.expectErr && tc.expectFound {
					s.Require().Equal(expectedValue.String(), tc.actualResultProto.String())
				}
			}
		})
	}
}

// TestMustSet tests that MustSet updates the store correctly
// and panics if an error is encountered.
func (s *TestSuite) TestMustSet() {
	tests := map[string]struct {
		// keys and values to preset
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
		"basic valid AuthParams test": {
			setKey: keyA,
			setValue: &authtypes.Params{
				MaxMemoCharacters: 600,
			},

			actualResultProto: &authtypes.Params{},
		},
		"invalid set value": {
			setKey:   keyA,
			setValue: (*sdk.DecProto)(nil),

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
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
		// keys and values to preset
		preSetKeyValues map[string]sdk.Dec

		// keys and values to attempt to get and validate
		expectedGetKeyValues map[string]sdk.Dec

		expectPanic bool
	}{
		"valid get": {
			preSetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: sdk.OneDec().Add(sdk.OneDec()),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},

			expectedGetKeyValues: map[string]sdk.Dec{
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

			expectedGetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: {}, // this one panics
			},

			expectPanic: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			// Setup
			for key, value := range tc.preSetKeyValues {
				osmoutils.MustSetDec(s.store, []byte(key), value)
			}

			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				for key, expectedValue := range tc.expectedGetKeyValues {
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
// N.B.: It is non-trivial to cause a panic
// by calling `MustSetDec` because it provides
// a valid proto argument to `MustSet` which will
// only panic if the proto argument is invalid.
// Therefore, we only test a success case here.
func (s *TestSuite) TestMustSetDec() {
	originalDecValue := sdk.OneDec()

	// System under test.
	osmoutils.MustSetDec(s.store, []byte(keyA), originalDecValue)

	// Assertions.
	retrievedDecVaue := osmoutils.MustGetDec(s.store, []byte(keyA))
	s.Require().Equal(originalDecValue.String(), retrievedDecVaue.String())
}

func (s *TestSuite) TestHasAnyAtPrefix() {
	testcases := map[string]struct {
		// if prefix is set, startValue and endValue are ignored.
		// we either create an iterator prefix or a range iterator.
		prefix     string
		startValue string
		endValue   string
		preSetKeys []string
		isReverse  bool

		expectedValue bool
		expectedErr   error
	}{
		"has one": {
			preSetKeys: oneA,

			prefix: prefixOne,

			expectedValue: true,
		},
		"has multiple": {
			preSetKeys: oneABC,

			prefix: prefixOne,

			expectedValue: true,
		},
		"has none": {
			preSetKeys: oneABC,

			prefix: prefixTwo,

			expectedValue: false,
		},
		"prefix lexicogrpahically below existing - does not find correctly": {
			preSetKeys: twoAB,

			prefix: prefixOne,

			expectedValue: false,
		},
		"prefix lexicogrpahically above existing - does not find correctly": {
			preSetKeys: twoAB,

			prefix: string(sdk.PrefixEndBytes([]byte(prefixTwo))),

			expectedValue: false,
		},
		"parse with error": {
			preSetKeys: oneABC,

			prefix: prefixOne,

			expectedErr: mockError,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			for i, key := range tc.preSetKeys {
				s.store.Set([]byte(key), []byte(fmt.Sprintf("%v", i)))
			}

			mockParseValueFn := mockParseValue
			if tc.expectedErr != nil {
				mockParseValueFn = mockParseValueWithError
			}

			actualValue, err := osmoutils.HasAnyAtPrefix(s.store, []byte(tc.prefix), mockParseValueFn)

			if tc.expectedErr != nil {
				s.Require().ErrorContains(err, tc.expectedErr.Error())
				s.Require().False(actualValue)
				return
			}

			s.Require().NoError(err)

			s.Require().Equal(tc.expectedValue, actualValue)
		})
	}
}

func (s *TestSuite) TestGetDec() {
	tests := map[string]struct {
		// keys and values to preset
		preSetKeyValues map[string]sdk.Dec

		// keys and values to attempt to get and validate
		expectedGetKeyValues map[string]sdk.Dec

		expectError error
	}{
		"valid get": {
			preSetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: sdk.OneDec().Add(sdk.OneDec()),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},

			expectedGetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyB: sdk.OneDec().Add(sdk.OneDec()),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},
		},
		"error: attempt to get non-existent key": {
			preSetKeyValues: map[string]sdk.Dec{
				keyA: sdk.OneDec(),
				keyC: sdk.OneDec().Add(sdk.OneDec()).Add(sdk.OneDec()),
			},

			expectedGetKeyValues: map[string]sdk.Dec{
				keyB: {},
			},

			expectError: osmoutils.DecNotFoundError{Key: keyB},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			// Setup
			for key, value := range tc.preSetKeyValues {
				osmoutils.MustSetDec(s.store, []byte(key), value)
			}

			for key, expectedValue := range tc.expectedGetKeyValues {
				// System under test.
				actualDec, err := osmoutils.GetDec(s.store, []byte(key))

				// Assertions.

				if tc.expectError != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, tc.expectError)
					return
				}

				s.Require().NoError(err)
				s.Require().Equal(expectedValue.String(), actualDec.String())
			}
		})

	}
}
