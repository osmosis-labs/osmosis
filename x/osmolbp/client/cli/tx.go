package cli

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        osmolbp.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", osmolbp.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CreateLBPCmd(),
	)

	return cmd
}

// CreateLBPCmd broadcast MsgCreateLBP.
func CreateLBPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-lbp [flags]",
		Short: "create-lbp creates s lbp",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreateLBPMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateLBP())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagTokenIn)
	_ = cmd.MarkFlagRequired(FlagTokenOut)
	_ = cmd.MarkFlagRequired(FlagStartTime)
	_ = cmd.MarkFlagRequired(FlagDuration)
	_ = cmd.MarkFlagRequired(FlagInitialDeposit)

	return cmd
}

func NewBuildCreateLBPMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	tokenIn, err := fs.GetString(FlagTokenIn)
	if err != nil {
		return txf, nil, err
	}

	tokenOut, err := fs.GetString(FlagTokenOut)
	if err != nil {
		return txf, nil, err
	}
	startTimeStr, err := fs.GetString(FlagStartTime)
	if err != nil {
		return txf, nil, err
	}
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return txf, nil, fmt.Errorf("could not parse time: %w", err)
	}
	duration, err := fs.GetDuration(FlagDuration)
	if err != nil {
		return txf, nil, fmt.Errorf("could not parse time: %w", err)
	}
	InitialDepositStr, err := fs.GetString(FlagInitialDeposit)
	if err != nil {
		return txf, nil, err
	}
	InitialDeposit, err := sdk.ParseCoinNormalized(InitialDepositStr)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse Initial_deposit amoung: %s", InitialDepositStr)
	}
	treasuryStr, err := fs.GetString(FlagTreasury)
	if err != nil {
		return txf, nil, err
	}
	treasury, err := sdk.AccAddressFromBech32(treasuryStr)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse treasury address: %s", treasuryStr)
	}
	msg := &api.MsgCreateLBP{
		TokenIn: tokenIn,
		TokenOut: tokenOut,
		StartTime: startTime,
		Duration: duration,
		InitialDeposit: InitialDeposit,
		Treasury: treasury.String(),
	}
	return txf, msg, nil
}