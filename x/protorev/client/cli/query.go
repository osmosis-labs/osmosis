package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
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
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryMaxPoolPointsPerTxCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryMaxPoolPointsPerBlockCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryBaseDenomsCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryEnabledCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryInfoByPoolTypeCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryPoolCmd)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, NewQueryAllProtocolRevenueCmd)

	return cmd
}

// NewQueryParamsCmd returns the command to query the module params
func NewQueryParamsCmd() (*osmocli.QueryDescriptor, *types.QueryParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query the module params",
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
		Use:   "profits-by-denom",
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
		Use:                "statistics-by-route",
		Short:              "Query statistics about a specific arbitrage route",
		Long:               `{{.Short}}{{.ExampleHeader}}{{.CommandPrefix}} statistics-by-route [1,2,3] `,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{"Route": parseRoute},
	}, &types.QueryGetProtoRevStatisticsByRouteRequest{}
}

// NewQueryAllRouteStatisticsCmd returns the command to query all route statistics of protorev
func NewQueryAllRouteStatisticsCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevAllRouteStatisticsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-statistics",
		Short: "Query all ProtoRev statistics",
	}, &types.QueryGetProtoRevAllRouteStatisticsRequest{}
}

// NewQueryTokenPairArbRoutesCmd returns the command to query the token pair arb routes
func NewQueryTokenPairArbRoutesCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevTokenPairArbRoutesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "hot-routes",
		Short: "Query the ProtoRev hot routes currently being used",
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

// NewQueryMaxPoolPointsPerTxCmd returns the command to query the max pool points per tx
func NewQueryMaxPoolPointsPerTxCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevMaxPoolPointsPerTxRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "max-pool-points-per-tx",
		Short: "Query the max pool points per tx",
	}, &types.QueryGetProtoRevMaxPoolPointsPerTxRequest{}
}

// NewQueryMaxPoolPointsPerBlockCmd returns the command to query the max pool points per block
func NewQueryMaxPoolPointsPerBlockCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevMaxPoolPointsPerBlockRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "max-pool-points-per-block",
		Short: "Query the max pool points per block",
	}, &types.QueryGetProtoRevMaxPoolPointsPerBlockRequest{}
}

// NewQueryBaseDenomsCmd returns the command to query the base denoms
func NewQueryBaseDenomsCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevBaseDenomsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "base-denoms",
		Short: "Query the base denoms used to construct arbitrage routes",
	}, &types.QueryGetProtoRevBaseDenomsRequest{}
}

// NewQueryEnabled returns the command to query the enabled status of protorev
func NewQueryEnabledCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevEnabledRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "enabled",
		Short: "Query whether protorev is currently enabled",
	}, &types.QueryGetProtoRevEnabledRequest{}
}

// NewQueryInfoByPoolTypeCmd returns the command to query the pool type info of protorev
func NewQueryInfoByPoolTypeCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevInfoByPoolTypeRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "info-by-pool-type",
		Short: "Query the pool info used to determine how computationally expensive a route is",
	}, &types.QueryGetProtoRevInfoByPoolTypeRequest{}
}

// NewQueryPoolCmd returns the command to query the pool id for a given denom pair stored via the highest liquidity method in ProtoRev
func NewQueryPoolCmd() (*osmocli.QueryDescriptor, *types.QueryGetProtoRevPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool",
		Short: "Query the pool id for a given denom pair stored via the highest liquidity method in ProtoRev",
	}, &types.QueryGetProtoRevPoolRequest{}
}

// NewQueryAllProtocolRevenueCmd returns the command to query protocol revenue across all modules
func NewQueryAllProtocolRevenueCmd() (*osmocli.QueryDescriptor, *types.QueryGetAllProtocolRevenueRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-proto-rev",
		Short: "Query protocol revenue across all modules",
	}, &types.QueryGetAllProtocolRevenueRequest{}
}

// convert a string array "[1,2,3]" to []uint64
//
//nolint:unparam
func parseRoute(arg string, _ *pflag.FlagSet) (any, osmocli.FieldReadLocation, error) {
	var route []uint64
	err := json.Unmarshal([]byte(arg), &route)
	if err != nil {
		return nil, osmocli.UsedArg, err
	}
	return route, osmocli.UsedArg, err
}
