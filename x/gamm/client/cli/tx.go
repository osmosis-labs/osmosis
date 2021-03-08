package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/c-osmosis/osmosis/x/gamm/types"
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
		NewCreatePoolCmd(),
		NewJoinPoolCmd(),
		NewExitPoolCmd(),
		NewSwapExactAmountInCmd(),
		NewSwapExactAmountOutCmd(),
		NewJoinSwapExternAmountIn(),
		NewJoinSwapShareAmountOut(),
		NewExitSwapExternAmountOut(),
	)

	return txCmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool",
		Short: "create a new pool and provide the liquidity to it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreatePoolMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePool())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolRecordTokens)
	_ = cmd.MarkFlagRequired(FlagPoolRecordTokenWeights)
	_ = cmd.MarkFlagRequired(FlagSwapFee)
	_ = cmd.MarkFlagRequired(FlagExitFee)

	return cmd
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
	_ = cmd.MarkFlagRequired(FlagSwapRouteAmounts)

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

	cmd.Flags().AddFlagSet(FlagSetQuerySwapRoutes())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteAmounts)

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
		Args:  cobra.ExactArgs(2),
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

func NewBuildCreatePoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	recordTokenStrs, err := fs.GetStringArray(FlagPoolRecordTokens)
	if err != nil {
		return txf, nil, err
	}
	if len(recordTokenStrs) < 2 {
		return txf, nil, fmt.Errorf("bind tokens should be more than 2")
	}

	recordTokenWeightStrs, err := fs.GetStringArray(FlagPoolRecordTokenWeights)
	if err != nil {
		return txf, nil, err
	}
	if len(recordTokenStrs) != len(recordTokenWeightStrs) {
		return txf, nil, fmt.Errorf("tokens and token weights should have same length")
	}

	recordTokens := sdk.Coins{}
	for i := 0; i < len(recordTokenStrs); i++ {
		parsed, err := sdk.ParseCoinNormalized(recordTokenStrs[i])
		if err != nil {
			return txf, nil, err
		}
		recordTokens = append(recordTokens, parsed)
	}

	var recordWeights []sdk.Int
	for i := 0; i < len(recordTokenWeightStrs); i++ {
		parsed, ok := sdk.NewIntFromString(recordTokenWeightStrs[i])
		if !ok {
			return txf, nil, fmt.Errorf("invalid token weight (%s)", recordTokenWeightStrs[i])
		}
		recordWeights = append(recordWeights, parsed)
	}

	swapFeeStr, err := fs.GetString(FlagSwapFee)
	if err != nil {
		return txf, nil, err
	}
	swapFee, err := sdk.NewDecFromStr(swapFeeStr)
	if err != nil {
		return txf, nil, err
	}

	exitFeeStr, err := fs.GetString(FlagExitFee)
	if err != nil {
		return txf, nil, err
	}
	exitFee, err := sdk.NewDecFromStr(exitFeeStr)
	if err != nil {
		return txf, nil, err
	}

	var records []types.Record
	for i := 0; i < len(recordTokens); i++ {
		recordToken := recordTokens[i]
		recordWeight := recordWeights[i]

		record := types.Record{
			Weight: recordWeight,
			Token:  recordToken,
		}

		records = append(records, record)
	}

	msg := &types.MsgCreatePool{
		Sender: clientCtx.GetFromAddress().String(),
		PoolParams: types.PoolParams{
			Lock:    false,
			SwapFee: swapFee,
			ExitFee: exitFee,
		},
		Records: records,
	}

	return txf, msg, nil
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
		parsed, err := sdk.ParseCoinNormalized(maxAmountsInStrs[i])
		if err != nil {
			return txf, nil, err
		}
		maxAmountsIn = append(maxAmountsIn, parsed)
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
		parsed, err := sdk.ParseCoinNormalized(minAmountsOutStrs[i])
		if err != nil {
			return txf, nil, err
		}
		minAmountsOut = append(minAmountsOut, parsed)
	}

	msg := &types.MsgExitPool{
		Sender:        clientCtx.GetFromAddress().String(),
		PoolId:        poolId,
		ShareInAmount: shareAmountIn,
		TokenOutMins:  minAmountsOut,
	}

	return txf, msg, nil
}

func swapAmountInRoutes(fs *flag.FlagSet) ([]types.SwapAmountInRoute, error) {
	swapRoutePoolIds, err := fs.GetStringArray(FlagSwapRoutePoolIds)
	if err != nil {
		return nil, err
	}

	swapRouteAmounts, err := fs.GetStringArray(FlagSwapRouteAmounts)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteAmounts) {
		return nil, errors.New("swap route pool ids and amounts mismatch")
	}

	routes := []types.SwapAmountInRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountInRoute{
			PoolId:        uint64(pID),
			TokenOutDenom: swapRouteAmounts[index],
		})
	}
	return routes, nil
}

func swapAmountOutRoutes(fs *flag.FlagSet) ([]types.SwapAmountOutRoute, error) {
	swapRoutePoolIds, err := fs.GetStringArray(FlagSwapRoutePoolIds)
	if err != nil {
		return nil, err
	}

	swapRouteAmounts, err := fs.GetStringArray(FlagSwapRouteAmounts)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteAmounts) {
		return nil, errors.New("swap route pool ids and amounts mismatch")
	}

	routes := []types.SwapAmountOutRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountOutRoute{
			PoolId:       uint64(pID),
			TokenInDenom: swapRouteAmounts[index],
		})
	}
	return routes, nil
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
		return txf, nil, errors.New("invalid token out min amount")
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
		return txf, nil, errors.New("share out amount")
	}

	shareInMaxAmt, ok := sdk.NewIntFromString(shareInMaxAmtStr)
	if !ok {
		return txf, nil, errors.New("token in max amount")
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
