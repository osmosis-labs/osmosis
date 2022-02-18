package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
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
		NewExitSwapShareAmountIn(),
	)

	return txCmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [flags]",
		Short: "create a new pool and provide the liquidity to it",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new pool and provide the liquidity to it.
Pool initialization parameters must be provided through a pool JSON file.

Example:
$ %s tx gamm create-pool --pool-file="path/to/pool.json" --from mykey

Where pool.json contains:
{
	"weights": "4uatom,4osmo,2uakt",
	"initial-deposit": "100uatom,5osmo,20uakt",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h"
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

			txf, msg, err := NewBuildCreateBalancerPoolMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePool())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolFile)

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

func NewBuildCreateBalancerPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {

	pool, err := parseCreatePoolFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse pool: %w", err)
	}

	deposit, err := sdk.ParseCoinsNormalized(pool.InitialDeposit)
	if err != nil {
		return txf, nil, err
	}

	poolAssetCoins, err := sdk.ParseDecCoins(pool.Weights)
	if err != nil {
		return txf, nil, err
	}

	if len(deposit) != len(poolAssetCoins) {
		return txf, nil, errors.New("deposit tokens and token weights should have same length")
	}

	swapFee, err := sdk.NewDecFromStr(pool.SwapFee)
	if err != nil {
		return txf, nil, err
	}

	exitFee, err := sdk.NewDecFromStr(pool.ExitFee)
	if err != nil {
		return txf, nil, err
	}

	var poolAssets []types.PoolAsset
	for i := 0; i < len(poolAssetCoins); i++ {

		if poolAssetCoins[i].Denom != deposit[i].Denom {
			return txf, nil, errors.New("deposit tokens and token weights should have same denom order")
		}

		poolAssets = append(poolAssets, types.PoolAsset{
			Weight: poolAssetCoins[i].Amount.RoundInt(),
			Token:  deposit[i],
		})
	}

	poolParams := &balancer.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}

	msg := &balancer.MsgCreateBalancerPool{
		Sender:             clientCtx.GetFromAddress().String(),
		PoolParams:         poolParams,
		PoolAssets:         poolAssets,
		FuturePoolGovernor: pool.FutureGovernor,
	}

	if (pool.SmoothWeightChangeParams != smoothWeightChangeParamsInputs{}) {
		duration, err := time.ParseDuration(pool.SmoothWeightChangeParams.Duration)
		if err != nil {
			return txf, nil, fmt.Errorf("could not parse duration: %w", err)
		}

		targetPoolAssetCoins, err := sdk.ParseDecCoins(pool.SmoothWeightChangeParams.TargetPoolWeights)
		if err != nil {
			return txf, nil, err
		}

		var targetPoolAssets []types.PoolAsset
		for i := 0; i < len(targetPoolAssetCoins); i++ {

			if targetPoolAssetCoins[i].Denom != poolAssetCoins[i].Denom {
				return txf, nil, errors.New("initial pool weights and target pool weights should have same denom order")
			}

			targetPoolAssets = append(targetPoolAssets, types.PoolAsset{
				Weight: targetPoolAssetCoins[i].Amount.RoundInt(),
				Token:  deposit[i],
				// TODO: This doesn't make sense. Should only use denom, not an sdk.Coin
			})
		}

		smoothWeightParams := balancer.SmoothWeightChangeParams{
			Duration:           duration,
			InitialPoolWeights: poolAssets,
			TargetPoolWeights:  targetPoolAssets,
		}

		if pool.SmoothWeightChangeParams.StartTime != "" {
			startTime, err := time.Parse(time.RFC3339, pool.SmoothWeightChangeParams.StartTime)
			if err != nil {
				return txf, nil, fmt.Errorf("could not parse time: %w", err)
			}

			smoothWeightParams.StartTime = startTime
		}

		msg.PoolParams.SmoothWeightChangeParams = &smoothWeightParams
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

	swapRouteDenoms, err := fs.GetStringArray(FlagSwapRouteDenoms)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteDenoms) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountInRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountInRoute{
			PoolId:        uint64(pID),
			TokenOutDenom: swapRouteDenoms[index],
		})
	}
	return routes, nil
}

func swapAmountOutRoutes(fs *flag.FlagSet) ([]types.SwapAmountOutRoute, error) {
	swapRoutePoolIds, err := fs.GetStringArray(FlagSwapRoutePoolIds)
	if err != nil {
		return nil, err
	}

	swapRouteDenoms, err := fs.GetStringArray(FlagSwapRouteDenoms)
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIds) != len(swapRouteDenoms) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountOutRoute{}
	for index, poolIDStr := range swapRoutePoolIds {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountOutRoute{
			PoolId:       uint64(pID),
			TokenInDenom: swapRouteDenoms[index],
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
