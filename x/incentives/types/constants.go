package types

import time "time"

var (
	BaseGasFeeForAddRewardToGauge = 10_000
	// We set the default value to 1ns, as this is the only uptime we support as long as charging is disabled (or
	// until more supported uptimes are authorized by governance).
	DefaultConcentratedUptime = time.Nanosecond
)
