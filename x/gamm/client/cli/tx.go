package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Generalized automated market maker transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewJoinPoolCmd(),
		NewExitPoolCmd(),
		NewJoinSwapExternAmountIn(),
		NewJoinSwapShareAmountOut(),
		NewExitSwapExternAmountOut(),
		NewExitSwapShareAmountIn(),
	)

	return txCmd
}

func NewJoinPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-pool",
		Short: "join a new pool and provide the liquidity to it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildJoinPoolMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetJoinPool())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagShareAmountOut)
	_ = cmd.MarkFlagRequired(FlagMaxAmountsIn)

	return cmd
}

func NewExitPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exit-pool",
		Short: "exit a new pool and withdraw the liquidity from it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildExitPoolMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetExitPool())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagShareAmountIn)
	_ = cmd.MarkFlagRequired(FlagMinAmountsOut)

	return cmd
}

func NewJoinSwapExternAmountIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-swap-extern-amount-in [token-in] [share-out-min-amount]",
		Short: "join swap extern amount in",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildJoinSwapExternAmountInMsg(clientCtx, args[0], args[1], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetJoinSwapExternAmount())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

func NewJoinSwapShareAmountOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-swap-share-amount-out [token-in-denom] [token-in-max-amount] [share-out-amount]",
		Short: "join swap share amount out",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildJoinSwapShareAmountOutMsg(clientCtx, args[0], args[1], args[2], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetJoinSwapExternAmount())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

func NewExitSwapExternAmountOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exit-swap-extern-amount-out [token-out] [share-in-max-amount]",
		Short: "exit swap extern amount out",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildExitSwapExternAmountOutMsg(clientCtx, args[0], args[1], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetJoinSwapExternAmount())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

func NewExitSwapShareAmountIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exit-swap-share-amount-in [token-out-denom] [share-in-amount] [token-out-min-amount]",
		Short: "exit swap share amount in",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildExitSwapShareAmountInMsg(clientCtx, args[0], args[1], args[2], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetJoinSwapExternAmount())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagPoolId)

	return cmd
}

func NewBuildJoinPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	shareAmountOutStr, err := fs.GetString(FlagShareAmountOut)
	if err != nil {
		return txf, nil, err
	}

	shareAmountOut, ok := sdk.NewIntFromString(shareAmountOutStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid share amount out")
	}

	maxAmountsInStrs, err := fs.GetStringArray(FlagMaxAmountsIn)
	if err != nil {
		return txf, nil, err
	}

	maxAmountsIn := sdk.Coins{}
	for i := 0; i < len(maxAmountsInStrs); i++ {
		parsed, err := sdk.ParseCoinsNormalized(maxAmountsInStrs[i])
		if err != nil {
			return txf, nil, err
		}
		maxAmountsIn = maxAmountsIn.Add(parsed...)
	}

	msg := &types.MsgJoinPool{
		Sender:         clientCtx.GetFromAddress().String(),
		PoolId:         poolId,
		ShareOutAmount: shareAmountOut,
		TokenInMaxs:    maxAmountsIn,
	}

	return txf, msg, nil
}

func NewBuildExitPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	shareAmountInStr, err := fs.GetString(FlagShareAmountIn)
	if err != nil {
		return txf, nil, err
	}

	shareAmountIn, ok := sdk.NewIntFromString(shareAmountInStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid share amount in")
	}

	minAmountsOutStrs, err := fs.GetStringArray(FlagMinAmountsOut)
	if err != nil {
		return txf, nil, err
	}

	minAmountsOut := sdk.Coins{}
	for i := 0; i < len(minAmountsOutStrs); i++ {
		parsed, err := sdk.ParseCoinsNormalized(minAmountsOutStrs[i])
		if err != nil {
			return txf, nil, err
		}
		minAmountsOut = minAmountsOut.Add(parsed...)
	}

	msg := &types.MsgExitPool{
		Sender:        clientCtx.GetFromAddress().String(),
		PoolId:        poolId,
		ShareInAmount: shareAmountIn,
		TokenOutMins:  minAmountsOut,
	}

	return txf, msg, nil
}

func NewBuildJoinSwapExternAmountInMsg(clientCtx client.Context, tokenInStr, shareOutMinAmountStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolID, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
	if err != nil {
		return txf, nil, err
	}

	shareOutMinAmount, ok := sdk.NewIntFromString(shareOutMinAmountStr)
	if !ok {
		return txf, nil, errors.New("invalid share out min amount")
	}
	msg := &types.MsgJoinSwapExternAmountIn{
		Sender:            clientCtx.GetFromAddress().String(),
		PoolId:            poolID,
		TokenIn:           tokenIn,
		ShareOutMinAmount: shareOutMinAmount,
	}

	return txf, msg, nil
}

func NewBuildJoinSwapShareAmountOutMsg(clientCtx client.Context, tokenInDenom, tokenInMaxAmtStr, shareOutAmtStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolID, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	tokenInMaxAmt, ok := sdk.NewIntFromString(tokenInMaxAmtStr)
	if !ok {
		return txf, nil, errors.New("token in max amount")
	}

	shareOutAmt, ok := sdk.NewIntFromString(shareOutAmtStr)
	if !ok {
		return txf, nil, errors.New("share out amount")
	}

	msg := &types.MsgJoinSwapShareAmountOut{
		Sender:           clientCtx.GetFromAddress().String(),
		PoolId:           poolID,
		TokenInDenom:     tokenInDenom,
		TokenInMaxAmount: tokenInMaxAmt,
		ShareOutAmount:   shareOutAmt,
	}

	return txf, msg, nil
}

func NewBuildExitSwapExternAmountOutMsg(clientCtx client.Context, tokenOutStr, shareInMaxAmtStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolID, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
	if err != nil {
		return txf, nil, errors.New("token out")
	}

	shareInMaxAmt, ok := sdk.NewIntFromString(shareInMaxAmtStr)
	if !ok {
		return txf, nil, errors.New("share in max amount")
	}

	msg := &types.MsgExitSwapExternAmountOut{
		Sender:           clientCtx.GetFromAddress().String(),
		PoolId:           poolID,
		TokenOut:         tokenOut,
		ShareInMaxAmount: shareInMaxAmt,
	}

	return txf, msg, nil
}

func NewBuildExitSwapShareAmountInMsg(clientCtx client.Context, tokenOutDenom, shareInAmtStr, tokenOutMinAmountStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolID, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	shareInAmt, ok := sdk.NewIntFromString(shareInAmtStr)
	if !ok {
		return txf, nil, errors.New("share in amount")
	}

	tokenOutMinAmount, ok := sdk.NewIntFromString(tokenOutMinAmountStr)
	if !ok {
		return txf, nil, errors.New("token out min amount")
	}

	msg := &types.MsgExitSwapShareAmountIn{
		Sender:            clientCtx.GetFromAddress().String(),
		PoolId:            poolID,
		TokenOutDenom:     tokenOutDenom,
		ShareInAmount:     shareInAmt,
		TokenOutMinAmount: tokenOutMinAmount,
	}

	return txf, msg, nil
}
