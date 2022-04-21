package types

import (
	fmt "fmt"
	"runtime"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
	// new epoch is next block of epoch end block
	BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
}

var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiEpochHooks []EpochHooks

func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the number of epoch that is ending.
func (h MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].AfterEpochEnd(ctx, epochIdentifier, epochNumber)
	}
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].BeforeEpochStart(ctx, epochIdentifier, epochNumber)
	}
}

func panicCatchingEpochHook(
	ctx sdk.Context,
	hookFn func(ctx sdk.Context, epochIdentifier string, epochNumber int64),
	epochIdentifier string,
	epochNumber int64) {

	defer func() {
		if recovErr := recover(); recovErr != nil {
			fmt.Println("Recovering from panic:", recovErr)
			fmt.Println("Stack Trace:")
			debug.PrintStack()
			switch e := recovErr.(type) {
			case string:
				ctx.Logger().Error("Recovering from panicrecovered (string) panic:", e)
			case runtime.Error:
				fmt.Println("recovered (runtime.Error) panic:", e.Error())
			case error:
				fmt.Println("recovered (error) panic:", e.Error())
			default:
				fmt.Println("recovered (default) panic:", e)
			}
			fmt.Println(string(debug.Stack()))
			return
		}
	}()
}
