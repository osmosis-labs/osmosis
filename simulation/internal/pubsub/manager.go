package pubsub

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/multierr"

	"github.com/osmosis-labs/osmosis/v10/simulation/simtypes"
)

var _ simtypes.PubSubManager = &Manager{}

type Manager struct {
	sim           *simtypes.SimCtx
	subscriptions map[string][]callbackfnWithMetadata
}

type callbackfnWithMetadata struct {
	subscriberName string
	callback       simtypes.SimCallbackFn
}

func NewPubSubManager(sim *simtypes.SimCtx) Manager {
	return Manager{sim: sim, subscriptions: map[string][]callbackfnWithMetadata{}}
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

func (m *Manager) Publish(ctx sdk.Context, key string, value interface{}) error {
	subscriptions, ok := m.subscriptions[key]
	if !ok {
		return nil
	}
	var result error
	for _, s := range subscriptions {
		err := s.callback(m.sim, ctx, value)
		result = multierr.Append(result, err)
	}
	return result
}
