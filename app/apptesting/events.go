package apptesting

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"slices"
)

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events
// of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	s.Require().Equal(numEventsExpected, len(actualEvents))
}

func (s *KeeperTestHelper) FindEvent(events []abci.Event, name string) abci.Event {
	index := slices.IndexFunc(events, func(e abci.Event) bool { return e.Type == name })
	if index == -1 {
		return abci.Event{}
	}
	return events[index]
}

func (s *KeeperTestHelper) ExtractAttributes(event abci.Event) map[string]string {
	attrs := make(map[string]string)
	if event.Attributes == nil {
		return attrs
	}
	for _, a := range event.Attributes {
		attrs[a.Key] = a.Value
	}
	return attrs
}
