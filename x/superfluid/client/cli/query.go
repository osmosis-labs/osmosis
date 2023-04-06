package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdAllSuperfluidAssets(),
		GetCmdAssetMultiplier(),
		GetCmdAllIntermediaryAccounts(),
		GetCmdConnectedIntermediaryAccount(),
		GetCmdSuperfluidDelegationAmount(),
		GetCmdSuperfluidDelegationsByDelegator(),
		GetCmdSuperfluidUndelegationsByDelegator(),
		GetCmdTotalSuperfluidDelegations(),
		GetCmdTotalDelegationByDelegator(),
		GetCmdUnpoolWhitelist(),
	)

	return cmd
}

// GetCmdQueryParams implements a command to fetch superfluid parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current superfluid parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query parameters for the superfluid module:

$ <appd> query superfluid params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			// NOTE: THIS IS NON-STANDARD, SO WE HAVE TO THINK ABOUT BREAKING IT
			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdAllSuperfluidAssets() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AllAssetsRequest](
		"all-superfluid-assets",
		"Query all superfluid assets", "",
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdAssetMultiplier() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AssetMultiplierRequest](
		"asset-multiplier [denom]",
		"Query asset multiplier by denom",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} asset-multiplier gamm/pool/1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdAllIntermediaryAccounts() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AllIntermediaryAccountsRequest](
		"all-intermediary-accounts",
		"Query all superfluid intermediary accounts",
		`{{.Short}}`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdConnectedIntermediaryAccount() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.ConnectedIntermediaryAccountRequest](
		"connected-intermediary-account [lock_id]",
		"Query connected intermediary account",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} connected-intermediary-account 1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdSuperfluidDelegationAmount returns the coins superfluid delegated for a
// delegator, validator, denom.
func GetCmdSuperfluidDelegationAmount() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.SuperfluidDelegationAmountRequest](
		"superfluid-delegation-amount [delegator_address] [validator_address] [denom]",
		"Query coins superfluid delegated for a delegator, validator, denom", "",
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdSuperfluidDelegationsByDelegator returns the coins superfluid delegated for the specified delegator.
func GetCmdSuperfluidDelegationsByDelegator() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.SuperfluidDelegationAmountRequest](
		"superfluid-delegation-by-delegator [delegator_address]",
		"Query coins superfluid delegated for the specified delegator", "",
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdSuperfluidUndelegationsByDelegator returns the coins superfluid undelegated for the specified delegator.
func GetCmdSuperfluidUndelegationsByDelegator() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.SuperfluidUndelegationsByDelegatorRequest](
		"superfluid-undelegation-by-delegator [delegator_address]",
		"Query coins superfluid undelegated for the specified delegator", "",
		types.ModuleName, types.NewQueryClient,
	)
}

// GetCmdTotalSuperfluidDelegations returns total amount of base denom delegated via superfluid staking.
func GetCmdTotalSuperfluidDelegations() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.TotalSuperfluidDelegationsRequest](
		"total-superfluid-delegations",
		"Query total amount of osmo delegated via superfluid staking", "",
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdTotalDelegationByDelegator() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalDelegationByDelegatorRequest](
		"total-delegation-by-delegator [delegator_address]",
		"Query both superfluid delegation and normal delegation", "",
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdUnpoolWhitelist() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryUnpoolWhitelistRequest](
		"unpool-whitelist",
		"Query whitelisted pool ids to unpool", "",
		types.ModuleName, types.NewQueryClient,
	)
}
