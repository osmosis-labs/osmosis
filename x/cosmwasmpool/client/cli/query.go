package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		osmocli.GetParams[*queryproto.ParamsRequest](
			types.ModuleName, query.NewQueryClient),
	)

	return cmd
}
