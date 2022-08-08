package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/v10/x/launchpad"
	"github.com/osmosis-labs/osmosis/v10/x/launchpad/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
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
		CreateSaleCmd(),
		FinalizeSaleCmd(),
		SubscribeCmd(),
		WithdrawCmd(),
		ExitSaleCmd(),
	)

	return cmd
}

// CreateSaleCmd broadcast MsgCreateSale.
func CreateSaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `create [flags]`,
		Short: "Create or Setup a sale",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create or Setup a launchpad sale.
Example:
$ %s tx launchpad create --sale-file="path/to/sale.json [flags]

Sample sale.json file contents:
{
	"token-in": "token1",
	"token-out": "1000token2",
	"max-fee": ["100token1"],
	"start-time": "2022-06-02T11:18:11.000Z",
	"duration": "432000s",
	"recipient": "osmo1r85gjuck87f9hw7l2c30w3zh696xrq0lus0kq6"
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

			txf, msg, err := NewBuildCreateSaleMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateSale())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagSaleFile)

	return cmd
}

// finalizeSale broadcasts MsgFinalizeSale
func FinalizeSaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize [flags]",
		Short: "Finalize a sale",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Finalize a launchpad sale.
Example:
$ %s tx launchpad finalize --sale-id=1 [flags]
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

			txf, msg, err := NewBuildFinalizeSaleMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetFinalizeSale())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagSaleId)

	return cmd
}

// Subscribe broadcast MsgSubscribe.
func SubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscribe [flags]",
		Short:   "Subscribe or Join a sale",
		Example: "subscribe --sale-id=1 --amount=10 [flags]",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Subscribe or Join a launchpad sale.
Example:
$ %s tx launchpad subscribe --sale-id=1 --amount=10 (in the smallest denominator) [flags]
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

			txf, msg, err := NewBuildSubscribeMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetSubscribe())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagSaleId)
	_ = cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// SubscribeSale broadcast MsgSubscribe.
func WithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw [flags]",
		Short:   "Withdraw amount from a sale",
		Example: "withdraw --sale-id=1 --amount=10 [flags]",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw amount from a launchpad sale.
Example:
$ %s tx launchpad withdraw --sale-id=1 --amount=10 (in the smallest denominator) [flags]
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
			txf, msg, err := NewBuildWithdrawMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetWithdraw())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagSaleId)

	return cmd
}

// ExitSaleCmd broadcast MsgExitSale.
func ExitSaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exit --sale-id=sale-id [flags]",
		Short:   "Exit from a launchpad sale",
		Example: "exit --sale-id=1 [flags]",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Exit from a launchpad sale.
Example:
$ %s tx launchpad exit --sale-id=1 [flags]
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

			txf, msg, err := NewBuildExitSaleMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetExit())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagSaleId)

	return cmd
}

func NewBuildCreateSaleMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	s, err := parseCreateSaleFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse sale: %w", err)
	}
	m, err := s.ToMsgCreateSale(clientCtx.GetFromAddress().String())
	return txf, m, err
}

func NewBuildFinalizeSaleMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	saleId, err := fs.GetUint64(FlagSaleId)
	if err != nil {
		return txf, nil, err
	}

	msg := &types.MsgFinalizeSale{
		Sender: clientCtx.GetFromAddress().String(),
		SaleId: saleId,
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildSubscribeMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	saleId, err := fs.GetUint64(FlagSaleId)
	if err != nil {
		return txf, nil, err
	}

	amount, err := fs.GetInt64(FlagAmount)
	if err != nil {
		return txf, nil, err
	}
	msg := &types.MsgSubscribe{
		Sender: clientCtx.GetFromAddress().String(),
		SaleId: saleId,
		Amount: sdk.NewInt(amount),
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildWithdrawMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	saleId, err := fs.GetUint64(FlagSaleId)
	if err != nil {
		return txf, nil, err
	}
	amount, err := fs.GetInt64(FlagAmount)
	if err != nil {
		return txf, nil, err
	}
	msg := &types.MsgWithdraw{
		Sender: clientCtx.GetFromAddress().String(),
		SaleId: saleId,
	}
	if amount > 0 {
		amt := sdk.NewInt(amount)
		msg.Amount = &amt
	} else {
		msg.Amount = nil
	}

	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}

func NewBuildExitSaleMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	saleId, err := fs.GetUint64(FlagSaleId)
	if err != nil {
		return txf, nil, err
	}

	msg := &types.MsgExitSale{
		Sender: clientCtx.GetFromAddress().String(),
		SaleId: saleId,
	}
	if err = msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}
	return txf, msg, nil
}
