package cli

import (
	"errors"
	"fmt"

	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Swap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSwapExactAmountInCmd(),
		NewSwapExactAmountOutCmd(),
	)

	return txCmd
}

func NewSwapExactAmountInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-exact-amount-in [token-in] [token-out-min-amount]",
		Short: "swap exact amount in",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildSwapExactAmountInMsg(clientCtx, args[0], args[1], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetQuerySwapRoutes())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)

	return cmd
}

func NewSwapExactAmountOutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-exact-amount-out [token-out] [token-in-max-amount]",
		Short: "swap exact amount out",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildSwapExactAmountOutMsg(clientCtx, args[0], args[1], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetSwapAmountOutRoutes())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)

	return cmd
}

func NewBuildSwapExactAmountInMsg(clientCtx client.Context, tokenInStr, tokenOutMinAmtStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	routes, err := swapAmountInRoutes(fs)
	if err != nil {
		return txf, nil, err
	}

	tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
	if err != nil {
		return txf, nil, err
	}

	tokenOutMinAmt, ok := sdk.NewIntFromString(tokenOutMinAmtStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid token out min amount, %s", tokenOutMinAmtStr)
	}
	msg := &types.MsgSwapExactAmountIn{
		Sender:            clientCtx.GetFromAddress().String(),
		Routes:            routes,
		TokenIn:           tokenIn,
		TokenOutMinAmount: tokenOutMinAmt,
	}

	return txf, msg, nil
}

func NewBuildSwapExactAmountOutMsg(clientCtx client.Context, tokenOutStr, tokenInMaxAmountStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	routes, err := swapAmountOutRoutes(fs)
	if err != nil {
		return txf, nil, err
	}

	tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
	if err != nil {
		return txf, nil, err
	}

	tokenInMaxAmount, ok := sdk.NewIntFromString(tokenInMaxAmountStr)
	if !ok {
		return txf, nil, errors.New("invalid token in max amount")
	}
	msg := &types.MsgSwapExactAmountOut{
		Sender:           clientCtx.GetFromAddress().String(),
		Routes:           routes,
		TokenInMaxAmount: tokenInMaxAmount,
		TokenOut:         tokenOut,
	}

	return txf, msg, nil
}
