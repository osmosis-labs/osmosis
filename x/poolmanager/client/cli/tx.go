package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)

	osmocli.AddTxCmd(txCmd, NewSwapExactAmountInCmd)
	osmocli.AddTxCmd(txCmd, NewSwapExactAmountOutCmd)
	osmocli.AddTxCmd(txCmd, NewSplitRouteSwapExactAmountIn)
	osmocli.AddTxCmd(txCmd, NewSplitRouteSwapExactAmountOut)
	txCmd.AddCommand(NewSetDenomPairTakerFeeCmd())

	txCmd.AddCommand(
		NewCreatePoolCmd(),
	)

	return txCmd
}

func NewSwapExactAmountInCmd() (*osmocli.TxCliDesc, *types.MsgSwapExactAmountIn) {
	return &osmocli.TxCliDesc{
		Use:     "swap-exact-amount-in",
		Short:   "swap exact amount in",
		Example: "osmosisd tx poolmanager swap-exact-amount-in 2000000uosmo 1 --swap-route-pool-ids 5 --swap-route-denoms uion --from val --keyring-backend test -b=block --chain-id=localosmosis --fees 10000uosmo",
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
		Example:          "osmosisd tx poolmanager swap-exact-amount-out 100uion 1000000 --swap-route-pool-ids 1 --swap-route-denoms uosmo --from val --keyring-backend test -b=block --chain-id=localosmosis --fees 10000uosmo",
		NumArgs:          2,
		ParseAndBuildMsg: NewBuildSwapExactAmountOutMsg,
		Flags:            osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetMultihopSwapRoutes()}},
	}, &types.MsgSwapExactAmountOut{}
}

func NewSplitRouteSwapExactAmountIn() (*osmocli.TxCliDesc, *types.MsgSplitRouteSwapExactAmountIn) {
	return &osmocli.TxCliDesc{
		Use:   "split-route-swap-exact-amount-in",
		Short: "split route swap exact amount in",
		Example: `osmosisd tx poolmanager split-route-swap-exact-amount-in uosmo 1 --routes-file="./routes.json" --from val --keyring-backend test -b=block --chain-id=localosmosis --fees 10000uosmo
		- routes.json
		{
			"Route": [
			  {
			  "swap_amount_in_route": [
				{
				"pool_id": 1,
				"token_out_denom": "uion"
				},
				{
				"pool_id": 2,
				"token_out_denom": "uosmo"
				}
			  ],
			  "token_in_amount": 1000
			  },
			  {
			  "swap_amount_in_route": [
				{
				"pool_id": 3,
				"token_out_denom": "bar"
				},
				{
				"pool_id": 4,
				"token_out_denom": "uosmo"
				}
			  ],
			  "token_in_amount": 999
			  }
			]
		}
		`,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"Routes": osmocli.FlagOnlyParser(NewMsgNewSplitRouteSwapExactAmountIn),
		},
		Flags: osmocli.FlagDesc{
			RequiredFlags: []*flag.FlagSet{FlagSetCreateRoutes()},
		},
	}, &types.MsgSplitRouteSwapExactAmountIn{}
}

func NewSplitRouteSwapExactAmountOut() (*osmocli.TxCliDesc, *types.MsgSplitRouteSwapExactAmountOut) {
	return &osmocli.TxCliDesc{
		Use:   "split-route-swap-exact-amount-out",
		Short: "split route swap exact amount out",
		Example: `osmosisd tx poolmanager split-route-swap-exact-amount-out uosmo 1 --routes-file="./routes.json" --from val --keyring-backend test -b=block --chain-id=localosmosis --fees 10000uosmo
		- routes.json
		{
			"route": [
				{
				"swap_amount_out_route": [
					{
					"pool_id": 1,
					"token_in_denom": "uion"
					},
					{
					"pool_id": 2,
					"token_in_denom": "uosmo"
					}
				],
				"token_out_amount": 1000
				},
				{
				"swap_amount_out_route": [
					{
					"pool_id": 3,
					"token_in_denom": "uion"
					},
					{
					"pool_id": 4,
					"token_in_denom": "uosmo"
					}
				],
				"token_out_amount": 999
				}
			]
			}
		`,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"Routes": osmocli.FlagOnlyParser(NewMsgNewSplitRouteSwapExactAmountOut),
		},
		Flags: osmocli.FlagDesc{
			RequiredFlags: []*flag.FlagSet{FlagSetCreateRoutes()},
		},
	}, &types.MsgSplitRouteSwapExactAmountOut{}
}

func NewMsgNewSplitRouteSwapExactAmountOut(fs *flag.FlagSet) ([]types.SwapAmountOutSplitRoute, error) {
	routesFile, _ := fs.GetString(FlagRoutesFile)
	if routesFile == "" {
		return nil, fmt.Errorf("must pass in a routes json using the --%s flag", FlagRoutesFile)
	}

	contents, err := os.ReadFile(routesFile)
	if err != nil {
		return nil, err
	}

	var splitRouteJSONdata RoutesOut
	err = json.Unmarshal(contents, &splitRouteJSONdata)
	if err != nil {
		return nil, err
	}

	var splitRouteProto []types.SwapAmountOutSplitRoute
	for _, route := range splitRouteJSONdata.Route {
		protoRoute := types.SwapAmountOutSplitRoute{
			TokenOutAmount: osmomath.NewInt(route.TokenOutAmount),
		}
		protoRoute.Pools = append(protoRoute.Pools, route.Pools...)
		splitRouteProto = append(splitRouteProto, protoRoute)
	}

	return splitRouteProto, nil
}

func NewMsgNewSplitRouteSwapExactAmountIn(fs *flag.FlagSet) ([]types.SwapAmountInSplitRoute, error) {
	routesFile, _ := fs.GetString(FlagRoutesFile)
	if routesFile == "" {
		return nil, fmt.Errorf("must pass in a routes json using the --%s flag", FlagRoutesFile)
	}

	contents, err := os.ReadFile(routesFile)
	if err != nil {
		return nil, err
	}

	var splitRouteJSONdata RoutesIn
	err = json.Unmarshal(contents, &splitRouteJSONdata)
	if err != nil {
		return nil, err
	}

	var splitRouteProto []types.SwapAmountInSplitRoute
	for _, route := range splitRouteJSONdata.Route {
		protoRoute := types.SwapAmountInSplitRoute{
			TokenInAmount: osmomath.NewInt(route.TokenInAmount),
		}
		protoRoute.Pools = append(protoRoute.Pools, route.Pools...)
		splitRouteProto = append(splitRouteProto, protoRoute)
	}

	return splitRouteProto, nil
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

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [flags]",
		Short: "create a new pool and provide the liquidity to it",
		Long:  `Must provide path to a pool JSON file (--pool-file) describing the pool to be created`,
		Example: `Sample pool JSON file contents:
{
	"weights": "4uatom,4osmo,2uakt",
	"initial-deposit": "100uatom,5osmo,20uakt",
	"swap-fee": "0.01",
	"exit-fee": "0.01",
	"future-governor": "168h"
}
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

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

func NewBuildCreateBalancerPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	pool, err := parseCreateBalancerPoolFlags(fs)
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

	spreadFactor, err := osmomath.NewDecFromStr(pool.SwapFee)
	if err != nil {
		return txf, nil, err
	}

	exitFee, err := osmomath.NewDecFromStr(pool.ExitFee)
	if err != nil {
		return txf, nil, err
	}

	var poolAssets []balancer.PoolAsset
	for i := 0; i < len(poolAssetCoins); i++ {
		if poolAssetCoins[i].Denom != deposit[i].Denom {
			return txf, nil, errors.New("deposit tokens and token weights should have same denom order")
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
			return txf, nil, fmt.Errorf("could not parse duration: %w", err)
		}

		targetPoolAssetCoins, err := sdk.ParseDecCoins(pool.SmoothWeightChangeParams.TargetPoolWeights)
		if err != nil {
			return txf, nil, err
		}

		var targetPoolAssets []balancer.PoolAsset
		for i := 0; i < len(targetPoolAssetCoins); i++ {
			if targetPoolAssetCoins[i].Denom != poolAssetCoins[i].Denom {
				return txf, nil, errors.New("initial pool weights and target pool weights should have same denom order")
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
				return txf, nil, fmt.Errorf("could not parse time: %w", err)
			}

			smoothWeightParams.StartTime = startTime
		}

		msg.PoolParams.SmoothWeightChangeParams = &smoothWeightParams
	}

	return txf, msg, nil
}

// Apologies to whoever has to touch this next, this code is horrendous
func NewBuildCreateStableswapPoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	flags, err := parseCreateStableswapPoolFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse pool: %w", err)
	}

	deposit, err := ParseCoinsNoSort(flags.InitialDeposit)
	if err != nil {
		return txf, nil, err
	}

	spreadFactor, err := osmomath.NewDecFromStr(flags.SwapFee)
	if err != nil {
		return txf, nil, err
	}

	exitFee, err := osmomath.NewDecFromStr(flags.ExitFee)
	if err != nil {
		return txf, nil, err
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
				return txf, nil, err
			}
			scalingFactors = append(scalingFactors, u)
		}
		if len(scalingFactors) != len(deposit) {
			return txf, nil, fmt.Errorf("number of scaling factors doesn't match number of assets")
		}
	}

	msg := &stableswap.MsgCreateStableswapPool{
		Sender:                  clientCtx.GetFromAddress().String(),
		PoolParams:              poolParams,
		InitialPoolLiquidity:    deposit,
		ScalingFactors:          scalingFactors,
		ScalingFactorController: flags.ScalingFactorController,
		FuturePoolGovernor:      flags.FutureGovernor,
	}

	return txf, msg, nil
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

// NewCmdHandleDenomPairTakerFeeProposal implements a command handler for denom pair taker fee proposal
func NewCmdHandleDenomPairTakerFeeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-pair-taker-fee-proposal [denom-pairs-with-taker-fee] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a denom pair taker fee proposal",
		Long: strings.TrimSpace(`Submit a denom pair taker fee proposal.

Passing in denom-pairs-with-taker-fee separated by commas would be parsed automatically to pairs of denomPairTakerFee records.
Ex) denom-pair-taker-fee-proposal uion,uosmo,0.0016,stake,uosmo,0.005,uatom,uosmo,0.0015 ->
[uion<>uosmo, takerFee 0.16%]
[stake<>uosmo, takerFee 0.5%]
[uatom<>uosmo, removes from state since its being set to the default takerFee value]

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseDenomPairTakerFeeArgToContent(cmd, args[0])
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

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)

	return cmd
}

func NewSetDenomPairTakerFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-denom-pair-taker-fee [flags]",
		Short: "allows admin addresses to set the taker fee for a denom pair",
		Long: strings.TrimSpace(`Allows admin addresses to set the taker fee for a denom pair.

Passing in set-denom-pair-taker-fee separated by commas would be parsed automatically to pairs of denomPairTakerFee records.
Ex) set-denom-pair-taker-fee uion,uosmo,0.0016,stake,uosmo,0.005,uatom,uosmo,0.0015 ->

[uion->uosmo, takerFee 0.16%]
[stake->uosmo, takerFee 0.5%]
[uatom->uosmo, removes from state since its being set to the default takerFee value]

NOTE: Denom pair taker fees are now uni-directional, so if you want a new taker fee to be charged in both directions, you need to set two records.
In other words, to set a taker fee for uosmo<->uion, you need to set it as follows:
set-denom-pair-taker-fee uosmo,uion,0.0016,uion,uosmo,0.0016

		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := parseDenomPairTakerFeeArgToMsg(clientCtx, args[0])
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePool())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseDenomPairTakerFeeArgToContent(cmd *cobra.Command, arg string) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	denomPairTakerFee, err := ParseDenomPairTakerFee(arg)
	if err != nil {
		return nil, err
	}

	content := &types.DenomPairTakerFeeProposal{
		Title:             title,
		Description:       description,
		DenomPairTakerFee: denomPairTakerFee,
	}

	return content, nil
}

func parseDenomPairTakerFeeArgToMsg(clientCtx client.Context, arg string) (sdk.Msg, error) {
	denomPairTakerFee, err := ParseDenomPairTakerFee(arg)
	if err != nil {
		return nil, err
	}

	msg := &types.MsgSetDenomPairTakerFee{
		Sender:            clientCtx.GetFromAddress().String(),
		DenomPairTakerFee: denomPairTakerFee,
	}

	return msg, nil
}

func ParseDenomPairTakerFee(arg string) ([]types.DenomPairTakerFee, error) {
	denomPairTakerFeeRecords := strings.Split(arg, ",")

	if len(denomPairTakerFeeRecords)%3 != 0 {
		return nil, fmt.Errorf("denomPairTakerFeeRecords must be a list of tokenInDenom, tokenOutDenom, and takerFee separated by commas")
	}

	finaldenomPairTakerFeeRecordsRecords := []types.DenomPairTakerFee{}
	i := 0
	for i < len(denomPairTakerFeeRecords) {
		tokenInDenom := denomPairTakerFeeRecords[i]
		tokenOutDenom := denomPairTakerFeeRecords[i+1]

		takerFeeStr := denomPairTakerFeeRecords[i+2]
		takerFee, err := osmomath.NewDecFromStr(takerFeeStr)
		if err != nil {
			return nil, err
		}

		finaldenomPairTakerFeeRecordsRecords = append(finaldenomPairTakerFeeRecordsRecords, types.DenomPairTakerFee{
			TokenInDenom:  tokenInDenom,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
		})

		// increase counter by the next 3
		i = i + 3
	}

	return finaldenomPairTakerFeeRecordsRecords, nil
}
