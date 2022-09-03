package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events
// of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int, expectedEvents sdk.Events, expectPass bool) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}

	if expectPass {
		s.Require().Equal(expectedEvents[0], actualEvents[0])
		s.Require().Equal(numEventsExpected, len(actualEvents))
	} else {
		s.Require().NotEqual(expectedEvents[0], actualEvents[0])
		s.Require().NotEqual(numEventsExpected, len(actualEvents))
	}
}
