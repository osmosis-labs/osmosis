package cli

import (
	"fmt"

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

	_ = cmd.MarkFlagRequired(FlagPoolBindTokens)
	_ = cmd.MarkFlagRequired(FlagPoolBindTokenWeights)
	_ = cmd.MarkFlagRequired(FlagSwapFee)

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
	_ = cmd.MarkFlagRequired(FlagPoolAmountOut)
	_ = cmd.MarkFlagRequired(FlagMaxAountsIn)

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
	_ = cmd.MarkFlagRequired(FlagPoolAmountIn)
	_ = cmd.MarkFlagRequired(FlagMinAmountsOut)

	return cmd
}

func NewBuildCreatePoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	bindTokenStrs, err := fs.GetStringArray(FlagPoolBindTokens)
	if err != nil {
		return txf, nil, err
	}
	if len(bindTokenStrs) < 2 {
		return txf, nil, fmt.Errorf("bind tokens should be more than 2")
	}

	bindTokenWeightStrs, err := fs.GetStringArray(FlagPoolBindTokenWeights)
	if err != nil {
		return txf, nil, err
	}
	if len(bindTokenStrs) != len(bindTokenWeightStrs) {
		return txf, nil, fmt.Errorf("bind tokens and token weight should have same length")
	}

	bindTokensSdk := sdk.Coins{}
	for i := 0; i < len(bindTokenStrs); i++ {
		parsed, err := sdk.ParseCoinNormalized(bindTokenStrs[i])
		if err != nil {
			return txf, nil, err
		}
		bindTokensSdk = append(bindTokensSdk, parsed)
	}

	var bindWeights []sdk.Dec
	for i := 0; i < len(bindTokenWeightStrs); i++ {
		parsed, err := sdk.NewDecFromStr(bindTokenWeightStrs[i])
		if err != nil {
			return txf, nil, err
		}
		bindWeights = append(bindWeights, parsed)
	}

	swapFeeStr, err := fs.GetString(FlagSwapFee)
	if err != nil {
		return txf, nil, err
	}
	swapFee, err := sdk.NewDecFromStr(swapFeeStr)
	if err != nil {
		return txf, nil, err
	}

	customDenom, err := fs.GetString(FlagPoolTokenCustomDenom)
	if err != nil {
		return txf, nil, err
	}

	description, err := fs.GetString(FlagPoolTokenDescription)
	if err != nil {
		return txf, nil, err
	}

	var bindTokens []types.BindTokenInfo
	for i := 0; i < len(bindTokensSdk); i++ {
		bindTokenSdk := bindTokensSdk[i]

		bindToken := types.BindTokenInfo{
			Denom:  bindTokenSdk.Denom,
			Weight: bindWeights[i],
			Amount: bindTokenSdk.Amount,
		}

		bindTokens = append(bindTokens, bindToken)
	}

	msg := &types.MsgCreatePool{
		Sender:  clientCtx.GetFromAddress(),
		SwapFee: swapFee,
		LpToken: types.LPTokenInfo{
			Denom:       customDenom,
			Description: description,
		},
		BindTokens: bindTokens,
	}

	return txf, msg, nil
}

func NewBuildJoinPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	poolAmountOutStr, err := fs.GetString(FlagPoolAmountOut)
	if err != nil {
		return txf, nil, err
	}

	poolAmountOut, ok := sdk.NewIntFromString(poolAmountOutStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid pool amount out")
	}

	maxAmountsInStrs, err := fs.GetStringArray(FlagMaxAountsIn)
	if err != nil {
		return txf, nil, err
	}

	maxAountsInSdk := sdk.Coins{}
	for i := 0; i < len(maxAmountsInStrs); i++ {
		parsed, err := sdk.ParseCoinNormalized(maxAmountsInStrs[i])
		if err != nil {
			return txf, nil, err
		}
		maxAountsInSdk = append(maxAountsInSdk, parsed)
	}

	var maxAmountsIn []types.MaxAmountIn
	for i := 0; i < len(maxAountsInSdk); i++ {
		maxAmountInSdk := maxAountsInSdk[i]

		maxAmountIn := types.MaxAmountIn{
			Denom:     maxAmountInSdk.Denom,
			MaxAmount: maxAmountInSdk.Amount,
		}

		maxAmountsIn = append(maxAmountsIn, maxAmountIn)
	}

	msg := &types.MsgJoinPool{
		Sender:        clientCtx.GetFromAddress(),
		TargetPoolId:  poolId,
		PoolAmountOut: poolAmountOut,
		MaxAmountsIn:  maxAmountsIn,
	}

	return txf, msg, nil
}

func NewBuildExitPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := fs.GetUint64(FlagPoolId)
	if err != nil {
		return txf, nil, err
	}

	poolAmountInStr, err := fs.GetString(FlagPoolAmountIn)
	if err != nil {
		return txf, nil, err
	}

	poolAmountIn, ok := sdk.NewIntFromString(poolAmountInStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid pool amount in")
	}

	minAmountsOutStrs, err := fs.GetStringArray(FlagMinAmountsOut)
	if err != nil {
		return txf, nil, err
	}

	minAountsOutSdk := sdk.Coins{}
	for i := 0; i < len(minAmountsOutStrs); i++ {
		parsed, err := sdk.ParseCoinNormalized(minAmountsOutStrs[i])
		if err != nil {
			return txf, nil, err
		}
		minAountsOutSdk = append(minAountsOutSdk, parsed)
	}

	var minAmountsOut []types.MinAmountOut
	for i := 0; i < len(minAountsOutSdk); i++ {
		minAmountOutSdk := minAountsOutSdk[i]

		minAmountOut := types.MinAmountOut{
			Denom:     minAmountOutSdk.Denom,
			MinAmount: minAmountOutSdk.Amount,
		}

		minAmountsOut = append(minAmountsOut, minAmountOut)
	}

	msg := &types.MsgExitPool{
		Sender:        clientCtx.GetFromAddress(),
		TargetPoolId:  poolId,
		PoolAmountIn:  poolAmountIn,
		MinAmountsOut: minAmountsOut,
	}

	return txf, msg, nil
}
