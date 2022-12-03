package osmocli

import (
	"fmt"
	"strconv"
	"time"
)

func ParseUint(arg string, fieldName string) (uint64, error) {
	v, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as uint for field %s: %w", arg, fieldName, err)
	}
	return v, nil
}

func ParseUnixTime(arg string, fieldName string) (time.Time, error) {
	timeUnix, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse %s as unix time for field %s: %w", arg, fieldName, err)
	}
	startTime := time.Unix(timeUnix, 0)
	return startTime, nil
}
