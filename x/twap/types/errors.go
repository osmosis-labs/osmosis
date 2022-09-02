package types

import (
	"fmt"
	time "time"
)

type EndTimeInFutureError struct {
	EndTime   time.Time
	BlockTime time.Time
}

func (e EndTimeInFutureError) Error() string {
	return fmt.Sprintf("called GetArithmeticTwap with an end time in the future."+
		" (end time %s, current time %s)", e.EndTime, e.BlockTime)
}

type StartTimeAfterEndTimeError struct {
	StartTime time.Time
	EndTime   time.Time
}

func (e StartTimeAfterEndTimeError) Error() string {
	return fmt.Sprintf("called GetArithmeticTwap with a start time that is after the end time."+
		" (start time %s, end time %s)", e.StartTime, e.EndTime)
}
