package valsetprefcli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	cmd.AddCommand(GetCmdValSetPref())
	return cmd
}

// GetCmdValSetPref takes the  address and returns the existing validator set for that address.
func GetCmdValSetPref() *cobra.Command {
	return osmocli.SimpleQueryCmd[*queryproto.UserValidatorPreferencesRequest](
		"val-set",
		"Query the validator set for a specific user address", "",
		types.ModuleName, queryproto.NewQueryClient,
	)
}
