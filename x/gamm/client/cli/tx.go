package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v21/x/gamm/types/migration"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewJoinPoolCmd)
	osmocli.AddTxCmd(txCmd, NewExitPoolCmd)
	osmocli.AddTxCmd(txCmd, NewSwapExactAmountInCmd)
	osmocli.AddTxCmd(txCmd, NewSwapExactAmountOutCmd)
	osmocli.AddTxCmd(txCmd, NewJoinSwapExternAmountIn)
	osmocli.AddTxCmd(txCmd, NewJoinSwapShareAmountOut)
	osmocli.AddTxCmd(txCmd, NewExitSwapExternAmountOut)
	osmocli.AddTxCmd(txCmd, NewExitSwapShareAmountIn)
	txCmd.AddCommand(
		NewCreatePoolCmd().BuildCommandCustomFn(),
		NewStableSwapAdjustScalingFactorsCmd(),
	)
	return txCmd
}

var poolIdFlagOverride = map[string]string{
	"poolid": FlagPoolId,
}

func NewCreatePoolCmd() *osmocli.TxCliDesc {
	desc := osmocli.TxCliDesc{
		Use:   "create-pool",
		Short: "create a new pool and provide the liquidity to it",
		Long: `Must provide path to a pool JSON file (--pool-file) describing the pool to be created
Sample pool JSON file contents for balancer:
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
	"exit-fee": "0.00",
	"future-governor": "168h",
	"scaling-factors": "1000,1"
}
`,
		NumArgs:          0,
		ParseAndBuildMsg: BuildCreatePoolCmd,
		Flags: osmocli.FlagDesc{
			RequiredFlags: []*flag.FlagSet{FlagSetCreatePoolFile()},
			OptionalFlags: []*flag.FlagSet{FlagSetCreatePoolType()},
		},
	}
	return &desc
}

func NewJoinPoolCmd() (*osmocli.TxCliDesc, *types.MsgJoinPool) {
	return &osmocli.TxCliDesc{
		Use:   "join-pool",
		Short: "join a new pool and provide the liquidity to it",
		CustomFlagOverrides: map[string]string{
			"poolid":         FlagPoolId,
			"ShareOutAmount": FlagShareAmountOut,
		},
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"TokenInMaxs": osmocli.FlagOnlyParser(maxAmountsInParser),
		},
		Flags: osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJoinPool()}},
	}, &types.MsgJoinPool{}
}

func NewExitPoolCmd() (*osmocli.TxCliDesc, *types.MsgExitPool) {
	return &osmocli.TxCliDesc{
		Use:   "exit-pool",
		Short: "exit a new pool and withdraw the liquidity from it",
		CustomFlagOverrides: map[string]string{
			"poolid":        FlagPoolId,
			"ShareInAmount": FlagShareAmountIn,
		},
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"TokenOutMins": osmocli.FlagOnlyParser(minAmountsOutParser),
		},
		Flags: osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetExitPool()}},
	}, &types.MsgExitPool{}
}

func NewSwapExactAmountInCmd() (*osmocli.TxCliDesc, *types.MsgSwapExactAmountIn) {
	return &osmocli.TxCliDesc{
		Use:   "swap-exact-amount-in",
		Short: "swap exact amount in",
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"Routes": osmocli.FlagOnlyParser(swapAmountInRoutes),
		},
		Flags: osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
	}, &types.MsgSwapExactAmountIn{}
}

func NewSwapExactAmountOutCmd() (*osmocli.TxCliDesc, *types.MsgSwapExactAmountOut) {
	// Can't get rid of this parser without a break, because the args are out of order.
	return &osmocli.TxCliDesc{
		Use:              "swap-exact-amount-out",
		Short:            "swap exact amount out",
		NumArgs:          2,
		ParseAndBuildMsg: NewBuildSwapExactAmountOutMsg,
		Flags:            osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
	}, &types.MsgSwapExactAmountOut{}
}

func NewJoinSwapExternAmountIn() (*osmocli.TxCliDesc, *types.MsgJoinSwapExternAmountIn) {
	return &osmocli.TxCliDesc{
		Use:                 "join-swap-extern-amount-in",
		Short:               "join swap extern amount in",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgJoinSwapExternAmountIn{}
}

func NewJoinSwapShareAmountOut() (*osmocli.TxCliDesc, *types.MsgJoinSwapShareAmountOut) {
	return &osmocli.TxCliDesc{
		Use:                 "join-swap-share-amount-out",
		Short:               "join swap share amount out",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgJoinSwapShareAmountOut{}
}

func NewExitSwapExternAmountOut() (*osmocli.TxCliDesc, *types.MsgExitSwapExternAmountOut) {
	return &osmocli.TxCliDesc{
		Use:                 "exit-swap-extern-amount-out",
		Short:               "exit swap extern amount out",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgExitSwapExternAmountOut{}
}

func NewExitSwapShareAmountIn() (*osmocli.TxCliDesc, *types.MsgExitSwapShareAmountIn) {
	return &osmocli.TxCliDesc{
		Use:                 "exit-swap-share-amount-in",
		Short:               "exit swap share amount in",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgExitSwapShareAmountIn{}
}

// TODO: Change these flags to args. Required flags don't make that much sense.
func NewStableSwapAdjustScalingFactorsCmd() *cobra.Command {
	cmd := osmocli.TxCliDesc{
		Use:              "adjust-scaling-factors --pool-id=[pool-id]  --scaling-factors=[scaling-factors]",
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

// NewCmdSubmitReplaceMigrationRecordsProposal implements a command handler for replace migration records proposal
func NewCmdSubmitReplaceMigrationRecordsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace-migration-records-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a replace migration record proposal",
		Long: strings.TrimSpace(`Submit a replace migration record proposal.

Passing in poolIds separated by commas would be parsed automatically to pairs of migration record.
Ex) 2,4,1,5 -> [(Balancer 2, CL 4), (Balancer 1, CL 5)]


		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseReplaceMigrationRecordsArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}
			if err = proposalMsg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagMigrationRecords, "", "The migration records array")

	return cmd
}

// NewCmdSubmitUpdateMigrationRecordsProposal implements a command handler for update migration records proposal
func NewCmdSubmitUpdateMigrationRecordsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-migration-records-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a update migration record proposal",
		Long: strings.TrimSpace(`Submit a update migration record proposal.

Passing in poolIds separated by commas would be parsed automatically to pairs of migration record.
Ex) 2,4,1,5 -> [(Balancer 2, CL 4), (Balancer 1, CL 5)]

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseUpdateMigrationRecordsArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}
			if err = proposalMsg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagMigrationRecords, "", "The migration records array")

	return cmd
}

// NewCmdSubmitUpdateMigrationRecordsProposal implements a command handler for update migration records proposal
func NewCmdSubmitCreateCLPoolAndLinkToCFMMProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-cl-pool-and-cfmm-link [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a create clpool and link to cfmm proposal",
		Long:  strings.TrimSpace(`submit a proposal to create CL pool and link to Balancer pool.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseCreateConcentratedLiquidityPoolArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}
			if err = proposalMsg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagPoolRecords, "", "The pool records array")

	return cmd
}

// NewCmdSubmitSetScalingFactorControllerProposal implements a command handler for the set scaling factor controller proposal
func NewCmdSubmitSetScalingFactorControllerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-scaling-factor-controller-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a set scaling factor controller proposal",
		Long: strings.TrimSpace(`Submit a set scaling factor controller proposal.

Sample proposal file:
{
	"title": "Set Scaling Factor Controller Proposal",
	"description": "Change scaling factor controller address from osmoXXX to osmoYYY"
	"pool-id": 1,
	"controller-address": "osmoYYY"
}
>>> osmosisd tx gov submit-proposal set-scaling-factor-controller-proposal \
        --proposal proposal.json \
		--deposit 1600000000uosmo \

Sample proposal with flags
>>> osmosisd tx gov submit-proposal set-scaling-factor-controller-proposal \
        --title "Set Scaling Factor Controller Proposal" \
		--summary "Change scaling factor controller address from osmoXXX to osmoYYY"
		--deposit 1600000000uosmo
		--pool-id 1
		--controller-address osmoYYY
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseSetScalingFactorControllerArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}
			if err = proposalMsg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().Uint64(FlagPoolId, 0, "stableswap pool-id")
	cmd.Flags().String(FlagScalingFactorControllerAddress, "", "target scaling factor controller address")

	return cmd
}

func BuildCreatePoolCmd(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	poolType, err := fs.GetString(FlagPoolType)
	if err != nil {
		return nil, err
	}
	poolType = strings.ToLower(poolType)

	var msg sdk.Msg
	if poolType == "balancer" || poolType == "uniswap" {
		msg, err = NewBuildCreateBalancerPoolMsg(clientCtx, fs)
		if err != nil {
			return nil, err
		}
	} else if poolType == "stableswap" {
		msg, err = NewBuildCreateStableswapPoolMsg(clientCtx, fs)
		if err != nil {
			return nil, err
		}
	}
	return msg, nil
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

	spreadFactor, err := osmomath.NewDecFromStr(pool.SwapFee)
	if err != nil {
		return nil, err
	}

	exitFee, err := osmomath.NewDecFromStr(pool.ExitFee)
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
		SwapFee: spreadFactor,
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

		if len(targetPoolAssetCoins) != len(poolAssetCoins) {
			return nil, errors.New("initial pool weights and target pool weights should have same length")
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

	spreadFactor, err := osmomath.NewDecFromStr(flags.SwapFee)
	if err != nil {
		return nil, err
	}

	exitFee, err := osmomath.NewDecFromStr(flags.ExitFee)
	if err != nil {
		return nil, err
	}

	poolParams := &stableswap.PoolParams{
		SwapFee: spreadFactor,
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

func maxAmountsInParser(fs *flag.FlagSet) (sdk.Coins, error) {
	return stringArrayCoinsParser(FlagMaxAmountsIn, fs)
}

func minAmountsOutParser(fs *flag.FlagSet) (sdk.Coins, error) {
	return stringArrayCoinsParser(FlagMinAmountsOut, fs)
}

func stringArrayCoinsParser(flagName string, fs *flag.FlagSet) (sdk.Coins, error) {
	amountsArr, err := fs.GetStringArray(flagName)
	if err != nil {
		return nil, err
	}

	coins := sdk.Coins{}
	for i := 0; i < len(amountsArr); i++ {
		parsed, err := sdk.ParseCoinsNormalized(amountsArr[i])
		if err != nil {
			return nil, err
		}
		coins = coins.Add(parsed...)
	}
	return coins, nil
}

func swapAmountInRoutes(fs *flag.FlagSet) ([]poolmanagertypes.SwapAmountInRoute, error) {
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

	routes := []poolmanagertypes.SwapAmountInRoute{}
	for index, poolIDStr := range swapRoutePoolIdsArray {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, poolmanagertypes.SwapAmountInRoute{
			PoolId:        uint64(pID),
			TokenOutDenom: swapRouteDenomsArray[index],
		})
	}
	return routes, nil
}

func swapAmountOutRoutes(fs *flag.FlagSet) ([]poolmanagertypes.SwapAmountOutRoute, error) {
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

	routes := []poolmanagertypes.SwapAmountOutRoute{}
	for index, poolIDStr := range swapRoutePoolIdsArray {
		pID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil, err
		}
		routes = append(routes, poolmanagertypes.SwapAmountOutRoute{
			PoolId:       uint64(pID),
			TokenInDenom: swapRouteDenomsArray[index],
		})
	}
	return routes, nil
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

	tokenInMaxAmount, ok := osmomath.NewIntFromString(tokenInMaxAmountStr)
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

func parseMigrationRecords(cmd *cobra.Command) ([]gammmigration.BalancerToConcentratedPoolLink, error) {
	assetsStr, err := cmd.Flags().GetString(FlagMigrationRecords)
	if err != nil {
		return nil, err
	}

	assets := strings.Split(assetsStr, ",")

	if len(assets)%2 != 0 {
		return nil, errors.New("migration records should be a list of balancer pool id and concentrated pool id pairs")
	}

	replaceMigrations := []gammmigration.BalancerToConcentratedPoolLink{}
	i := 0
	for i < len(assets) {
		balancerPoolId, err := strconv.Atoi(assets[i])
		if err != nil {
			return nil, err
		}
		clPoolId, err := strconv.Atoi(assets[i+1])
		if err != nil {
			return nil, err
		}

		replaceMigrations = append(replaceMigrations, gammmigration.BalancerToConcentratedPoolLink{
			BalancerPoolId: uint64(balancerPoolId),
			ClPoolId:       uint64(clPoolId),
		})

		// increase counter by the next 2
		i = i + 2
	}

	return replaceMigrations, nil
}

func parseReplaceMigrationRecordsArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	replaceMigrations, err := parseMigrationRecords(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.ReplaceMigrationRecordsProposal{
		Title:       title,
		Description: description,
		Records:     replaceMigrations,
	}
	return content, nil
}

func parseUpdateMigrationRecordsArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	replaceMigrations, err := parseMigrationRecords(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.UpdateMigrationRecordsProposal{
		Title:       title,
		Description: description,
		Records:     replaceMigrations,
	}
	return content, nil
}

func parseCreateConcentratedLiquidityPoolArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	poolRecordsWithCFMMLink, err := parsePoolRecordsWithCFMMLink(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal{
		Title:                   title,
		Description:             description,
		PoolRecordsWithCfmmLink: poolRecordsWithCFMMLink,
	}

	return content, nil
}

func parsePoolRecordsWithCFMMLink(cmd *cobra.Command) ([]types.PoolRecordWithCFMMLink, error) {
	poolRecordsStr, err := cmd.Flags().GetString(FlagPoolRecords)
	if err != nil {
		return nil, err
	}

	poolRecordsWithCFMMLink := strings.Split(poolRecordsStr, ",")

	if len(poolRecordsWithCFMMLink)%6 != 0 {
		return nil, fmt.Errorf("poolRecordswithCFMMLink must be a list of denom0, denom1, tickSpacing, exponentAtPriceOne, spreadFactor and balancerPoolId")
	}

	finalPoolRecords := []types.PoolRecordWithCFMMLink{}
	i := 0
	for i < len(poolRecordsWithCFMMLink) {
		denom0 := poolRecordsWithCFMMLink[i]
		denom1 := poolRecordsWithCFMMLink[i+1]

		tickSpacing, err := strconv.Atoi(poolRecordsWithCFMMLink[i+2])
		if err != nil {
			return nil, err
		}

		exponentAtPriceOneStr := poolRecordsWithCFMMLink[i+3]
		exponentAtPriceOne, ok := osmomath.NewIntFromString(exponentAtPriceOneStr)
		if !ok {
			return nil, fmt.Errorf("invalid exponentAtPriceOne: %s", exponentAtPriceOneStr)
		}

		spreadFactorStr := poolRecordsWithCFMMLink[i+4]
		spreadFactor, err := osmomath.NewDecFromStr(spreadFactorStr)
		if err != nil {
			return nil, err
		}

		balancerPoolId, err := strconv.Atoi(poolRecordsWithCFMMLink[i+5])
		if err != nil {
			return nil, err
		}

		finalPoolRecords = append(finalPoolRecords, types.PoolRecordWithCFMMLink{
			Denom0:             denom0,
			Denom1:             denom1,
			TickSpacing:        uint64(tickSpacing),
			ExponentAtPriceOne: exponentAtPriceOne,
			SpreadFactor:       spreadFactor,
			BalancerPoolId:     uint64(balancerPoolId),
		})

		// increase counter by the next 6
		i = i + 6
	}

	return finalPoolRecords, nil
}

func parseSetScalingFactorControllerArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	proposalFile, err := cmd.Flags().GetString(govcli.FlagProposal) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	if proposalFile != "" {
		contents, err := os.ReadFile(proposalFile)
		if err != nil {
			return nil, err
		}

		var proposal types.SetScalingFactorControllerProposal
		if err := json.Unmarshal(contents, &proposal); err != nil {
			return nil, err
		}
		return &proposal, nil
	}

	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	poolId, err := cmd.Flags().GetUint64(FlagPoolId)
	if err != nil {
		return nil, err
	}

	controllerAddress, err := cmd.Flags().GetString(FlagScalingFactorControllerAddress)
	if err != nil {
		return nil, err
	}

	content := &types.SetScalingFactorControllerProposal{
		Title:             title,
		Description:       description,
		PoolId:            poolId,
		ControllerAddress: controllerAddress,
	}

	return content, nil
}
