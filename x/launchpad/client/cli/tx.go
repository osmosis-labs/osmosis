package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/x/launchpad"
	"github.com/osmosis-labs/osmosis/x/launchpad/api"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"strings"
	"time"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        launchpad.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", launchpad.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CreateLBPCmd(),
		FinalizeLBPCmd(),
		SubscribeCmd(),
		WithdrawCmd(),
		ExitLBPCmd(),
	)

	return cmd
}

// CreateLBPCmd broadcast MsgCreateLBP.
func CreateLBPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Create or Setup LBP",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new LBP.

Example:
$ %s tx launchpad create --lbp-file="path/to/lbp.json" --from mykey

Where lbp.json contains:
{
	"token-in": "token1",
	"token-out": "token2",
	"initial-deposit": "1000token2",
	"start-time": "2022-05-23T11:17:36.755Z",
	"duration": 432000s,
	"treasury": "osmo1r85gjuck87f9hw7l2c30w3zh696xrq0lus0kq6"
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
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

	_ = cmd.MarkFlagRequired(FlagLBPFile)

	return cmd
}

// finalizeLBP broadcasts MsgFinalizeLBP
func FinalizeLBPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize [flags]",
		Short: "Finalize LBP",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildFinalizeLBPMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetFinalizeLBP())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

// Subscribe broadcast MsgSubscribe.
func SubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe [flags]",
		Short: "Subscribe or Join LBP",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildSubscribeMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetSubscribe())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// SubscribeLBP broadcast MsgSubscribe.
func WithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [flags]",
		Short: "Withdraw amount from LBP",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildWithdrawMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetWithdraw())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

// ExitLBPCmd broadcast MsgExitLBP.
func ExitLBPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exit [flags]",
		Short: "Exit from LBP",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildExitLBPMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetExit())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

func NewBuildCreateLBPMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	lbp, err := parseCreateLBPFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse lbp: %w", err)
	}

	InitialDeposit, err := sdk.ParseCoinNormalized(lbp.InitialDeposit)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse Initial-deposit amount: %s", lbp.InitialDeposit)
	}
	treasury, err := sdk.AccAddressFromBech32(lbp.Treasury)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse treasury address: %s", lbp.Treasury)
	}
	duration, err := time.ParseDuration(lbp.Duration)
	if err != nil {
		return txf, nil, err
	}

	msg := &api.MsgCreateLBP{
		TokenIn:        lbp.TokenIn,
		TokenOut:       lbp.TokenOut,
		StartTime:      lbp.StartTime,
		Duration:       duration,
		InitialDeposit: InitialDeposit,
		Treasury:       treasury.String(),
		Creator:        clientCtx.GetFromAddress().String(),
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	return txf, msg, nil
}

func NewBuildFinalizeLBPMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	msg := &api.MsgFinalizeLBP{
		Sender: clientCtx.GetFromAddress().String(),
		PoolId: poolId,
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildSubscribeMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	amount, err := fs.GetInt64(FlagAmount)
	if err != nil {
		return txf, nil, err
	}
	msg := &api.MsgSubscribe{
		Sender: clientCtx.GetFromAddress().String(),
		PoolId: poolId,
		Amount: sdk.NewInt(amount),
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildWithdrawMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	msg := &api.MsgWithdraw{
		Sender: clientCtx.GetFromAddress().String(),
		PoolId: poolId,
		Amount: nil,
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildExitLBPMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	msg := &api.MsgExitLBP{
		Sender: clientCtx.GetFromAddress().String(),
		PoolId: poolId,
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}
