package simtypes

import sdk "github.com/cosmos/cosmos-sdk/types"

type SimCallbackFn func(sim *SimCtx, ctx sdk.Context, value interface{}) error

type PubSubManager interface {
	Subscribe(key string, subName string, callback SimCallbackFn)
	Publish(sim *SimCtx, ctx sdk.Context, key string, value interface{}) error
}

type PropertyCheck interface {
	// A property check listens for signals on the listed channels, that the simulator can emit.
	// Known channel types right now:
	// * Post-Action execute
	// * Pre-Action execute
	// * Block end (can make listener execute every Nth block end)
	SubscriptionKeys() []string
	Check(sim *SimCtx, ctx sdk.Context, key string, value interface{}) error
}
