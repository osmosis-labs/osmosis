package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, RecoveredSinceQueryCmd)

	return cmd
}

func RecoveredSinceQueryCmd() (*osmocli.QueryDescriptor, *queryproto.RecoveredSinceDowntimeOfLengthRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "recovered-since downtime-duration recovery-duration",
		Short: "Queries if it has been at least <recovery-duration> since the chain was down for <downtime-duration>",
		Long: `{{.Short}}
downtime-duration is a duration, but is restricted to a smaller set. Heres a few from the set: 30s, 1m, 5m, 10m, 30m, 1h, 3 h, 6h, 12h, 24h, 36h, 48h]
{{.ExampleHeader}}
{{.CommandPrefix}} recovered-since 24h 30m`,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{"Downtime": parseDowntimeDuration},
	}, &queryproto.RecoveredSinceDowntimeOfLengthRequest{}
}

//nolint:unparam
func parseDowntimeDuration(arg string, _ *pflag.FlagSet) (any, osmocli.FieldReadLocation, error) {
	dur, err := time.ParseDuration(arg)
	if err != nil {
		return nil, osmocli.UsedArg, err
	}
	downtime, err := types.DowntimeByDuration(dur)
	return downtime, osmocli.UsedArg, err
}
