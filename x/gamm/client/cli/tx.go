package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"

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
		NewExitSwapShareAmountIn(),
		NewStableSwapAdjustScalingFactorsCmd(),
	)

	return txCmd
}

var poolIdFlagOverride = map[string]string{
	"poolid": FlagPoolId,
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [flags]",
		Short: "create a new pool and provide the liquidity to it",
		Long:  `Must provide path to a pool JSON file (--pool-file) describing the pool to be created`,
		Example: `Sample pool JSON file contents for balancer:
{
	"weights": "4uatom,4osmo,2uakt",
	"initial-deposit": "100uatom,5osmo,20uakt",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h"
}

For stableswap (demonstrating need for a 1:1000 scaling factor, see doc)
{
	"initial-deposit": "1000000uusdc,1000miliusdc",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h",
	"scaling-factors": "1000,1"
}
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
				WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			poolType, err := cmd.Flags().GetString(FlagPoolType)
			if err != nil {
				return err
			}
			poolType = strings.ToLower(poolType)

			var msg sdk.Msg
			if poolType == "balancer" || poolType == "uniswap" {
				msg, err = NewBuildCreateBalancerPoolMsg(clientCtx, cmd.Flags())
				if err != nil {
					return err
				}
			} else if poolType == "stableswap" {
				msg, err = NewBuildCreateStableswapPoolMsg(clientCtx, cmd.Flags())
				if err != nil {
					return err
				}
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
	cmd := osmocli.TxCliDesc{
		Use:              "join-pool",
		Short:            "join a new pool and provide the liquidity to it",
		NumArgs:          0,
		ParseAndBuildMsg: NewBuildJoinPoolMsg,
	}.BuildCommandCustomFn()

	cmd.Flags().AddFlagSet(FlagSetJoinPool())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagShareAmountOut)
	_ = cmd.MarkFlagRequired(FlagMaxAmountsIn)
	return cmd
}

func NewExitPoolCmd() *cobra.Command {
	cmd := osmocli.TxCliDesc{
		Use:              "exit-pool",
		Short:            "exit a new pool and withdraw the liquidity from it",
		NumArgs:          0,
		ParseAndBuildMsg: NewBuildExitPoolMsg,
	}.BuildCommandCustomFn()

	cmd.Flags().AddFlagSet(FlagSetExitPool())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagShareAmountIn)
	_ = cmd.MarkFlagRequired(FlagMinAmountsOut)
	return cmd
}

func NewSwapExactAmountInCmd() *cobra.Command {
	cmd := osmocli.TxCliDesc{
		Use:              "swap-exact-amount-in [token-in] [token-out-min-amount]",
		Short:            "swap exact amount in",
		NumArgs:          2,
		ParseAndBuildMsg: NewBuildSwapExactAmountInMsg,
	}.BuildCommandCustomFn()

	cmd.Flags().AddFlagSet(FlagSetQuerySwapRoutes())
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)
	return cmd
}

func NewSwapExactAmountOutCmd() *cobra.Command {
	cmd := osmocli.TxCliDesc{
		Use:              "swap-exact-amount-out [token-out] [token-in-max-amount]",
		Short:            "swap exact amount out",
		NumArgs:          2,
		ParseAndBuildMsg: NewBuildSwapExactAmountOutMsg,
	}.BuildCommandCustomFn()

	cmd.Flags().AddFlagSet(FlagSetSwapAmountOutRoutes())
	_ = cmd.MarkFlagRequired(FlagSwapRoutePoolIds)
	_ = cmd.MarkFlagRequired(FlagSwapRouteDenoms)
	return cmd
}

func NewJoinSwapExternAmountIn() *cobra.Command {
	cmd := osmocli.BuildTxCli[*types.MsgJoinSwapExternAmountIn](&osmocli.TxCliDesc{
		Use:                 "join-swap-extern-amount-in [token-in] [share-out-min-amount]",
		Short:               "join swap extern amount in",
		CustomFlagOverrides: poolIdFlagOverride,
	})

	cmd.Flags().AddFlagSet(FlagSetJustPoolId())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	return cmd
}

func NewJoinSwapShareAmountOut() *cobra.Command {
	cmd := osmocli.BuildTxCli[*types.MsgJoinSwapShareAmountOut](&osmocli.TxCliDesc{
		Use:                 "join-swap-share-amount-out [token-in-denom] [token-in-max-amount] [share-out-amount]",
		Short:               "join swap share amount out",
		CustomFlagOverrides: poolIdFlagOverride,
	})

	cmd.Flags().AddFlagSet(FlagSetJustPoolId())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	return cmd
}

func NewExitSwapExternAmountOut() *cobra.Command {
	cmd := osmocli.BuildTxCli[*types.MsgExitSwapExternAmountOut](&osmocli.TxCliDesc{
		Use:                 "exit-swap-extern-amount-out [token-out] [share-in-max-amount]",
		Short:               "exit swap extern amount out",
		CustomFlagOverrides: poolIdFlagOverride,
	})

	cmd.Flags().AddFlagSet(FlagSetJustPoolId())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	return cmd
}

func NewExitSwapShareAmountIn() *cobra.Command {
	cmd := osmocli.BuildTxCli[*types.MsgExitSwapShareAmountIn](&osmocli.TxCliDesc{
		Use:                 "exit-swap-share-amount-in [token-out-denom] [share-in-amount] [token-out-min-amount]",
		Short:               "exit swap share amount in",
		CustomFlagOverrides: poolIdFlagOverride,
	})

	cmd.Flags().AddFlagSet(FlagSetJustPoolId())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	return cmd
}

// TODO: Change these flags to args. Required flags don't make that much sense.
func NewStableSwapAdjustScalingFactorsCmd() *cobra.Command {
	cmd := osmocli.TxCliDesc{
		Use:              "adjust-scaling-factors --pool-id=[pool-id] --scaling-factors=[scaling-factors]",
		Short:            "adjust scaling factors",
		Example:          "osmosisd adjust-scaling-factors --pool-id=1 --scaling-factors=\"100, 100\"",
		NumArgs:          0,
		ParseAndBuildMsg: NewStableSwapAdjustScalingFactorsMsg,
	}.BuildCommandCustomFn()

	cmd.Flags().AddFlagSet(FlagSetAdjustScalingFactors())
	_ = cmd.MarkFlagRequired(FlagPoolId)
	_ = cmd.MarkFlagRequired(FlagScalingFactors)
	return cmd
}

func NewBuildCreateBalancerPoolMsg(clientCtx client.Context, fs *flag.FlagSet) (sdk.Msg, error) {
	pool, err := parseCreateBalancerPoolFlags(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool: %w", err)
	}

	deposit, err := sdk.ParseCoinsNormalized(pool.InitialDeposit)
	if err != nil {
		return nil, err
	}

	poolAssetCoins, err := sdk.ParseDecCoins(pool.Weights)
	if err != nil {
		return nil, err
	}

	if len(deposit) != len(poolAssetCoins) {
		return nil, errors.New("deposit tokens and token weights should have same length")
	}

	swapFee, err := sdk.NewDecFromStr(pool.SwapFee)
	if err != nil {
		return nil, err
	}

	exitFee, err := sdk.NewDecFromStr(pool.ExitFee)
	if err != nil {
		return nil, err
	}

	var poolAssets []balancer.PoolAsset
	for i := 0; i < len(poolAssetCoins); i++ {
		if poolAssetCoins[i].Denom != deposit[i].Denom {
			return nil, errors.New("deposit tokens and token weights should have same denom order")
		}

		poolAssets = append(poolAssets, balancer.PoolAsset{
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
			return nil, fmt.Errorf("could not parse duration: %w", err)
		}

		targetPoolAssetCoins, err := sdk.ParseDecCoins(pool.SmoothWeightChangeParams.TargetPoolWeights)
		if err != nil {
			return nil, err
		}

		var targetPoolAssets []balancer.PoolAsset
		for i := 0; i < len(targetPoolAssetCoins); i++ {
			if targetPoolAssetCoins[i].Denom != poolAssetCoins[i].Denom {
				return nil, errors.New("initial pool weights and target pool weights should have same denom order")
			}

			targetPoolAssets = append(targetPoolAssets, balancer.PoolAsset{
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
				return nil, fmt.Errorf("could not parse time: %w", err)
			}

			smoothWeightParams.StartTime = startTime
		}

		msg.PoolParams.SmoothWeightChangeParams = &smoothWeightParams
	}

	return msg, nil
}

// Apologies to whoever has to touch this next, this code is horrendous
func NewBuildCreateStableswapPoolMsg(clientCtx client.Context, fs *flag.FlagSet) (sdk.Msg, error) {
	flags, err := parseCreateStableswapPoolFlags(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool: %w", err)
	}

	deposit, err := ParseCoinsNoSort(flags.InitialDeposit)
	if err != nil {
		return nil, err
	}

	swapFee, err := sdk.NewDecFromStr(flags.SwapFee)
	if err != nil {
		return nil, err
	}

	exitFee, err := sdk.NewDecFromStr(flags.ExitFee)
	if err != nil {
		return nil, err
	}

	poolParams := &stableswap.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}

	scalingFactors := []uint64{}
	trimmedSfString := strings.Trim(flags.ScalingFactors, "[] {}")
	if len(trimmedSfString) > 0 {
		ints := strings.Split(trimmedSfString, ",")
		for _, i := range ints {
			u, err := strconv.ParseUint(i, 10, 64)
			if err != nil {
				return nil, err
			}
			scalingFactors = append(scalingFactors, u)
		}
		if len(scalingFactors) != len(deposit) {
			return nil, fmt.Errorf("number of scaling factors doesn't match number of assets")
		}
	}

	return &stableswap.MsgCreateStableswapPool{
		Sender:                  clientCtx.GetFromAddress().String(),
		PoolParams:              poolParams,
		InitialPoolLiquidity:    deposit,
		ScalingFactors:          scalingFactors,
		ScalingFactorController: flags.ScalingFactorController,
		FuturePoolGovernor:      flags.FutureGovernor,
	}, nil
}

func NewBuildJoinPoolMsg(clientCtx client.Context, _args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return nil, err
	}

	shareAmountOutStr, err := fs.GetString(FlagShareAmountOut)
	if err != nil {
		return nil, err
	}

	shareAmountOut, ok := sdk.NewIntFromString(shareAmountOutStr)
	if !ok {
		return nil, fmt.Errorf("invalid share amount out")
	}

	maxAmountsInStrs, err := fs.GetStringArray(FlagMaxAmountsIn)
	if err != nil {
		return nil, err
	}

	maxAmountsIn := sdk.Coins{}
	for i := 0; i < len(maxAmountsInStrs); i++ {
		parsed, err := sdk.ParseCoinsNormalized(maxAmountsInStrs[i])
		if err != nil {
			return nil, err
		}
		maxAmountsIn = maxAmountsIn.Add(parsed...)
	}

	return &types.MsgJoinPool{
		Sender:         clientCtx.GetFromAddress().String(),
		PoolId:         poolId,
		ShareOutAmount: shareAmountOut,
		TokenInMaxs:    maxAmountsIn,
	}, nil
}

func NewBuildExitPoolMsg(clientCtx client.Context, _args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return nil, err
	}

	shareAmountInStr, err := fs.GetString(FlagShareAmountIn)
	if err != nil {
		return nil, err
	}

	shareAmountIn, ok := sdk.NewIntFromString(shareAmountInStr)
	if !ok {
		return nil, fmt.Errorf("invalid share amount in")
	}

	minAmountsOutStrs, err := fs.GetStringArray(FlagMinAmountsOut)
	if err != nil {
		return nil, err
	}

	minAmountsOut := sdk.Coins{}
	for i := 0; i < len(minAmountsOutStrs); i++ {
		parsed, err := sdk.ParseCoinsNormalized(minAmountsOutStrs[i])
		if err != nil {
			return nil, err
		}
		minAmountsOut = minAmountsOut.Add(parsed...)
	}

	return &types.MsgExitPool{
		Sender:        clientCtx.GetFromAddress().String(),
		PoolId:        poolId,
		ShareInAmount: shareAmountIn,
		TokenOutMins:  minAmountsOut,
	}, nil
}

func swapAmountInRoutes(fs *flag.FlagSet) ([]types.SwapAmountInRoute, error) {
	swapRoutePoolIds, err := fs.GetString(FlagSwapRoutePoolIds)
	swapRoutePoolIdsArray := strings.Split(swapRoutePoolIds, ",")
	if err != nil {
		return nil, err
	}

	swapRouteDenoms, err := fs.GetString(FlagSwapRouteDenoms)
	swapRouteDenomsArray := strings.Split(swapRouteDenoms, ",")
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIdsArray) != len(swapRouteDenomsArray) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountInRoute{}
	for index, poolIDStr := range swapRoutePoolIdsArray {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountInRoute{
			PoolId:        uint64(pID),
			TokenOutDenom: swapRouteDenomsArray[index],
		})
	}
	return routes, nil
}

func swapAmountOutRoutes(fs *flag.FlagSet) ([]types.SwapAmountOutRoute, error) {
	swapRoutePoolIds, err := fs.GetString(FlagSwapRoutePoolIds)
	swapRoutePoolIdsArray := strings.Split(swapRoutePoolIds, ",")
	if err != nil {
		return nil, err
	}

	swapRouteDenoms, err := fs.GetString(FlagSwapRouteDenoms)
	swapRouteDenomsArray := strings.Split(swapRouteDenoms, ",")
	if err != nil {
		return nil, err
	}

	if len(swapRoutePoolIdsArray) != len(swapRouteDenomsArray) {
		return nil, errors.New("swap route pool ids and denoms mismatch")
	}

	routes := []types.SwapAmountOutRoute{}
	for index, poolIDStr := range swapRoutePoolIdsArray {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, types.SwapAmountOutRoute{
			PoolId:       uint64(pID),
			TokenInDenom: swapRouteDenomsArray[index],
		})
	}
	return routes, nil
}

func NewBuildSwapExactAmountInMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	tokenInStr, tokenOutMinAmtStr := args[0], args[1]
	routes, err := swapAmountInRoutes(fs)
	if err != nil {
		return nil, err
	}

	tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
	if err != nil {
		return nil, err
	}

	tokenOutMinAmt, ok := sdk.NewIntFromString(tokenOutMinAmtStr)
	if !ok {
		return nil, fmt.Errorf("invalid token out min amount, %s", tokenOutMinAmtStr)
	}
	return &types.MsgSwapExactAmountIn{
		Sender:            clientCtx.GetFromAddress().String(),
		Routes:            routes,
		TokenIn:           tokenIn,
		TokenOutMinAmount: tokenOutMinAmt,
	}, nil
}

func NewBuildSwapExactAmountOutMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	tokenOutStr, tokenInMaxAmountStr := args[0], args[1]
	routes, err := swapAmountOutRoutes(fs)
	if err != nil {
		return nil, err
	}

	tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
	if err != nil {
		return nil, err
	}

	tokenInMaxAmount, ok := sdk.NewIntFromString(tokenInMaxAmountStr)
	if !ok {
		return nil, errors.New("invalid token in max amount")
	}
	return &types.MsgSwapExactAmountOut{
		Sender:           clientCtx.GetFromAddress().String(),
		Routes:           routes,
		TokenInMaxAmount: tokenInMaxAmount,
		TokenOut:         tokenOut,
	}, nil
}

func NewStableSwapAdjustScalingFactorsMsg(clientCtx client.Context, _args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	poolID, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return nil, err
	}

	scalingFactorsStr, err := fs.GetString(FlagScalingFactors)
	if err != nil {
		return nil, err
	}

	scalingFactorsStrSlice := strings.Split(scalingFactorsStr, ",")

	scalingFactors := make([]uint64, len(scalingFactorsStrSlice))
	for i, scalingFactorStr := range scalingFactorsStrSlice {
		scalingFactor, err := strconv.ParseUint(scalingFactorStr, 10, 64)
		if err != nil {
			return nil, err
		}
		scalingFactors[i] = scalingFactor
	}

	msg := &stableswap.MsgStableSwapAdjustScalingFactors{
		Sender:         clientCtx.GetFromAddress().String(),
		PoolID:         poolID,
		ScalingFactors: scalingFactors,
	}

	return msg, nil
}

// ParseCoinsNoSort parses coins from coinsStr but does not sort them.
// Returns error if parsing fails.
func ParseCoinsNoSort(coinsStr string) (sdk.Coins, error) {
	coinStrs := strings.Split(coinsStr, ",")
	decCoins := make(sdk.DecCoins, len(coinStrs))
	for i, coinStr := range coinStrs {
		coin, err := sdk.ParseDecCoin(coinStr)
		if err != nil {
			return sdk.Coins{}, err
		}

		decCoins[i] = coin
	}
	return sdk.NormalizeCoins(decCoins), nil
}
