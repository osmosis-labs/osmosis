package cli

// import (
// 	"github.com/spf13/cobra"

// 	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
// 	"github.com/osmosis-labs/osmosis/v13/x/downtime-detector/client/queryproto"
// 	"github.com/osmosis-labs/osmosis/v13/x/downtime-detector/types"
// )

// // GetQueryCmd returns the cli query commands for this module.
// func GetQueryCmd() *cobra.Command {
// 	cmd := osmocli.QueryIndexCmd(types.ModuleName)
// 	osmocli.AddQueryCommand(cmd, RecoveredSinceQueryCmd)

// 	return cmd
// }

// // GetCmdLockedByID returns lock by id.
// func RecoveredSinceQueryCmd() (*osmocli.QueryDescriptor, *queryproto.RecoveredSinceDowntimeOfLengthRequest) {
// 	return &osmocli.QueryDescriptor{
// 		Use:   "recovered-since downtime-duration recovery-duration",
// 		Short: "Queries if it has been at least <recovery-duration> since the chain was down for <downtime-duration>",
// 		Long: `{{.Short}}
// downtime-duration is a duration, but is restricted to a smaller set. Heres a few from the set: 30s, 1m, 5m, 10m, 30m, 1h, 3 h, 6h, 12h, 24h, 36h, 48h]
// {{.ExampleHeader}}
// {{.CommandPrefix}} recovered-since 24h 30m`,
// 		QueryFnName: "LockedByID",
// 	}, &queryproto.RecoveredSinceDowntimeOfLengthRequest{}
// }
