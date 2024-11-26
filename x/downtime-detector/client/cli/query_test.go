package cli_test

import (
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

// We test the custom duration parser via this
func TestRecoveredSinceQueryCmd(t *testing.T) {
	desc, _ := cli.RecoveredSinceQueryCmd()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.RecoveredSinceDowntimeOfLengthRequest]{
		"basic test": {
			Cmd: "30s 10m",
			ExpectedQuery: &queryproto.RecoveredSinceDowntimeOfLengthRequest{
				Downtime: types.Downtime_DURATION_30S,
				Recovery: time.Minute * 10,
			},
		},
		"invalid duration": {
			Cmd: "31s 10m",
			ExpectedQuery: &queryproto.RecoveredSinceDowntimeOfLengthRequest{
				Downtime: types.Downtime_DURATION_30S,
				Recovery: time.Minute * 10,
			},
			ExpectedErr: true,
		},
		"90m": {
			Cmd: "90m 10m",
			ExpectedQuery: &queryproto.RecoveredSinceDowntimeOfLengthRequest{
				Downtime: types.Downtime_DURATION_1_5H,
				Recovery: time.Minute * 10,
			},
		},
		"1.5h": {
			Cmd: "1.5h 10m",
			ExpectedQuery: &queryproto.RecoveredSinceDowntimeOfLengthRequest{
				Downtime: types.Downtime_DURATION_1_5H,
				Recovery: time.Minute * 10,
			},
		},
		"1h30m": {
			Cmd: "1h30m 10m",
			ExpectedQuery: &queryproto.RecoveredSinceDowntimeOfLengthRequest{
				Downtime: types.Downtime_DURATION_1_5H,
				Recovery: time.Minute * 10,
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
