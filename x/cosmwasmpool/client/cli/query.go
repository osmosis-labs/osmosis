package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

// NewQueryCmd returns the cli query commands for this module.
func NewQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	qcGetter := queryproto.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdPools)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdContractInfoByPoolId)
	cmd.AddCommand(
		osmocli.GetParams[*queryproto.ParamsRequest](
			types.ModuleName, queryproto.NewQueryClient),
	)

	return cmd
}

func GetCmdPools() (*osmocli.QueryDescriptor, *queryproto.PoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`,
	}, &queryproto.PoolsRequest{}
}

func GetCmdContractInfoByPoolId() (*osmocli.QueryDescriptor, *queryproto.ContractInfoByPoolIdRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "contract-info",
		Short: "Query contract info by pool id",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`,
	}, &queryproto.ContractInfoByPoolIdRequest{}
}
