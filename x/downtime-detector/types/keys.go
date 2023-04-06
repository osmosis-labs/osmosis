package types

import fmt "fmt"

// There are few of these keys, so we don't concern ourselves with small key names.
var (
	lastBlockTimestampKey      = []byte("last_block_timestamp")
	lastDowntimeOfLengthPrefix = "last_downtime_of_length/%s"
)

func GetLastBlockTimestampKey() []byte { return lastBlockTimestampKey }

func GetLastDowntimeOfLengthKey(downtimeDur Downtime) []byte {
	return []byte(fmt.Sprintf(lastDowntimeOfLengthPrefix, downtimeDur.String()))
}
