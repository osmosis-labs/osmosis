package valsetprefcli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSetValSetCmd(),
	)

	return txCmd
}

func NewSetValSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-valset [delegator_addr] [validators] [weights]",
		Short:   "Creates a new validator set for the delegator with valOperAddress and weight",
		Example: "osmosisd tx valset-pref set-valset osmo1... osmovaloper1abc...,osmovaloper1def...  0.56,0.44",

		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valAddrs := osmoutils.ParseSdkValAddressFromString(args[1], ",")

			weights, err := osmoutils.ParseSdkDecFromString(args[2], ",")
			if err != nil {
				return err
			}

			if len(valAddrs) != len(weights) {
				return fmt.Errorf("the length of validator addresses and weights not matched")
			}

			if len(valAddrs) == 0 {
				return fmt.Errorf("records is empty")
			}

			var valset []types.ValidatorPreference
			for i, val := range valAddrs {
				valset = append(valset, types.ValidatorPreference{
					Weight:         weights[i],
					ValOperAddress: val.String(),
				})
			}

			msg := types.NewMsgSetValidatorSetPreference(
				delAddr,
				valset,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
