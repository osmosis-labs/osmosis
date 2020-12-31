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
	_ = cmd.MarkFlagRequired(FlagShareAmountIn)
	_ = cmd.MarkFlagRequired(FlagMinAmountsOut)

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

	maxAmountsInStrs, err := fs.GetStringArray(FlagMaxAountsIn)
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
