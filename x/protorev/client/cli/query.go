package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// NewCmdQuery returns the cli query commands for this module
func NewCmdQuery() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryParamsCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryNumberOfTradesCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryProfitsByDenomCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryAllProfitsCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryStatisticsByRouteCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryAllRouteStatisticsCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryTokenPairArbRoutesCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryAdminAccountCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryDeveloperAccountCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryMaxRoutesPerTxCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryMaxRoutesPerBlockCmd)

	return cmd
}

// NewQueryParamsCmd returns the command to query the current params
func NewQueryParamsCmd() (*osmocli.QueryDescriptor, *types.QueryParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query the current params",
	}, &types.QueryParamsRequest{}
}

// NewQueryNumberOfTradesCmd returns the command to query the number of trades executed by protorev
func NewQueryNumberOfTradesCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevNumberOfTradesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "number-of-trades",
		Short: "Query the number of cyclic arbitrage trades protorev has executed",
	}, &types.QueryGetProtoRevNumberOfTradesRequest{}
}

// NewQueryProfitsByDenomCmd returns the command to query the profits of protorev by denom
func NewQueryProfitsByDenomCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevProfitsByDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "profits-by-denom [denom]",
		Short: "Query the profits of protorev by denom",
		Long:  `{{.Short}}{{.ExampleHeader}}{{.CommandPrefix}} profits-by-denom uosmo`,
	}, &types.QueryGetProtoRevProfitsByDenomRequest{}
}

// NewQueryAllProfitsCmd returns the command to query all profits of protorev
func NewQueryAllProfitsCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevAllProfitsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-profits",
		Short: "Query all ProtoRev profits",
	}, &types.QueryGetProtoRevAllProfitsRequest{}
}

// NewQueryStatisticsByRoute returns the command to query the statistics of protorev by route
func NewQueryStatisticsByRouteCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevStatisticsByRouteRequest) {
	return &osmocli.QueryDescriptor{
		Use:                "statistics-by-route [route]",
		Short:              "Query the statistics of protorev by route",
		Long:               `{{.Short}}{{.ExampleHeader}}{{.CommandPrefix}} statistics-by-route [1,2,3]`,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{"Route": parseRoute},
	}, &types.QueryGetProtoRevStatisticsByRouteRequest{}
}

// NewQueryAllRouteStatisticsCmd returns the command to query all statistics of protorev
func NewQueryAllRouteStatisticsCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevAllRouteStatisticsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-statistics",
		Short: "Query all ProtoRev statistics",
	}, &types.QueryGetProtoRevAllRouteStatisticsRequest{}
}

// NewQueryTokenPairArbRoutesCmd returns the command to query the token pair arb routes
func NewQueryTokenPairArbRoutesCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevTokenPairArbRoutesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "token-pair-arb-routes",
		Short: "Query the ProtoRev token pair arb routes",
	}, &types.QueryGetProtoRevTokenPairArbRoutesRequest{}
}

// NewQueryAdminAccountCmd returns the command to query the admin account
func NewQueryAdminAccountCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevAdminAccountRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "admin-account",
		Short: "Query the admin account",
	}, &types.QueryGetProtoRevAdminAccountRequest{}
}

// NewQueryDeveloperAccountCmd returns the command to query the developer account
func NewQueryDeveloperAccountCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevDeveloperAccountRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "developer-account",
		Short: "Query the developer account",
	}, &types.QueryGetProtoRevDeveloperAccountRequest{}
}

// NewQueryMaxRoutesPerTxCmd returns the command to query the max routes per tx
func NewQueryMaxRoutesPerTxCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevMaxRoutesPerTxRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "max-routes-per-tx",
		Short: "Query the max routes per tx",
	}, &types.QueryGetProtoRevMaxRoutesPerTxRequest{}
}

// NewQueryMaxRoutesPerBlockCmd returns the command to query the max routes per block
func NewQueryMaxRoutesPerBlockCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevMaxRoutesPerBlockRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "max-routes-per-block",
		Short: "Query the max routes per block",
	}, &types.QueryGetProtoRevMaxRoutesPerBlockRequest{}
}

// convert a string array "[1,2,3]" to []uint64
func parseRoute(arg string, _ *pflag.FlagSet) (any, osmocli.FieldReadLocation, error) {
	var route []uint64
	err := json.Unmarshal([]byte(arg), &route)
	if err != nil {
		return nil, osmocli.UsedArg, err
	}
	return route, osmocli.UsedArg, err
}
