package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/epochs/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

// dummyEpochHook is a struct satisfying the epoch hook interface,
// that maintains a counter for how many times its been succesfully called,
// and a boolean for whether it should panic during its execution.
type dummyEpochHook struct {
	successCounter int
	shouldPanic    bool
}

func (hook *dummyEpochHook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if hook.shouldPanic {
		panic("dummyEpochHook is panicking")
	}
	hook.successCounter += 1
}

func (hook *dummyEpochHook) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if hook.shouldPanic {
		panic("dummyEpochHook is panicking")
	}
	hook.successCounter += 1
}

func (hook *dummyEpochHook) Clone() *dummyEpochHook {
	newHook := dummyEpochHook{shouldPanic: hook.shouldPanic, successCounter: hook.successCounter}
	return &newHook
}

var _ types.EpochHooks = &dummyEpochHook{}

func (suite *KeeperTestSuite) TestHooksPanicRecovery() {
	panicHook := dummyEpochHook{shouldPanic: true}
	noPanicHook := dummyEpochHook{shouldPanic: false}
	simpleHooks := []dummyEpochHook{panicHook, noPanicHook}

	tests := []struct {
		hooks                 []dummyEpochHook
		expectedCounterValues []int
	}{
		{[]dummyEpochHook{noPanicHook}, []int{1}},
		{simpleHooks, []int{0, 1}},
	}

	for tcIndex, tc := range tests {
		for epochActionSelector := 0; epochActionSelector < 2; epochActionSelector++ {
			suite.SetupTest()
			hookRefs := []types.EpochHooks{}

			for _, hook := range tc.hooks {
				hookRefs = append(hookRefs, hook.Clone())
			}

			hooks := types.NewMultiEpochHooks(hookRefs...)
			suite.NotPanics(func() {
				if epochActionSelector == 0 {
					hooks.BeforeEpochStart(suite.Ctx, "id", 0)
				} else if epochActionSelector == 1 {
					hooks.AfterEpochEnd(suite.Ctx, "id", 0)
				}
			})

			for i := 0; i < len(hooks); i++ {
				epochHook := hookRefs[i].(*dummyEpochHook)
				suite.Require().Equal(tc.expectedCounterValues[i], epochHook.successCounter, "test case index %d", tcIndex)
			}
		}
	}
}
