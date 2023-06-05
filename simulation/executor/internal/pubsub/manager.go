package pubsub

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/multierr"

	"github.com/osmosis-labs/osmosis/v16/simulation/simtypes"
)

var _ simtypes.PubSubManager = &Manager{}

type Manager struct {
	subscriptions map[string][]callbackfnWithMetadata
}

type callbackfnWithMetadata struct {
	subscriberName string
	callback       simtypes.SimCallbackFn
}

func NewPubSubManager() Manager {
	return Manager{subscriptions: map[string][]callbackfnWithMetadata{}}
}

func (m *Manager) Subscribe(key string, subName string, callback simtypes.SimCallbackFn) {
	subscriptions, ok := m.subscriptions[key]
	callbackStruct := callbackfnWithMetadata{subscriberName: subName, callback: callback}
	if ok {
		subscriptions = append(subscriptions, callbackStruct)
	} else {
		subscriptions = []callbackfnWithMetadata{callbackStruct}
	}
	m.subscriptions[key] = subscriptions
}

func (m *Manager) Publish(sim *simtypes.SimCtx, ctx sdk.Context, key string, value interface{}) error {
	subscriptions, ok := m.subscriptions[key]
	if !ok {
		return nil
	}
	var result error
	for _, s := range subscriptions {
		err := s.callback(sim, ctx, value)
		result = multierr.Append(result, err)
	}
	return result
}
