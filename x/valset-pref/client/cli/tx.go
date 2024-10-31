package valsetprefcli

import (
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

func GetTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewSetValSetCmd)
	osmocli.AddTxCmd(txCmd, NewDelValSetCmd)
	// TODO: Uncomment when undelegate is implemented
	// https://github.com/osmosis-labs/osmosis/issues/6686
	//osmocli.AddTxCmd(txCmd, NewUnDelValSetCmd)
	osmocli.AddTxCmd(txCmd, NewUndelRebalancedValSetCmd)
	osmocli.AddTxCmd(txCmd, NewReDelValSetCmd)
	osmocli.AddTxCmd(txCmd, NewWithRewValSetCmd)
	return txCmd
}

func NewSetValSetCmd() (*osmocli.TxCliDesc, *types.MsgSetValidatorSetPreference) {
	return &osmocli.TxCliDesc{
		Use:              "set-valset",
		Short:            "Creates a new validator set for the delegator with valOperAddress and weight",
		Example:          "osmosisd tx valset-pref set-valset osmo1... osmovaloper1abc...,osmovaloper1def...  0.56,0.44",
		NumArgs:          3,
		ParseAndBuildMsg: NewMsgSetValidatorSetPreference,
	}, &types.MsgSetValidatorSetPreference{}
}

func NewDelValSetCmd() (*osmocli.TxCliDesc, *types.MsgDelegateToValidatorSet) {
	return &osmocli.TxCliDesc{
		Use:     "delegate-valset",
		Short:   "Delegate tokens to existing valset using delegatorAddress and tokenAmount.",
		Example: "osmosisd tx valset-pref delegate-valset osmo1... 100stake",
		NumArgs: 2,
	}, &types.MsgDelegateToValidatorSet{}
}

// TODO: Uncomment when undelegate is implemented
// https://github.com/osmosis-labs/osmosis/issues/6686
// func NewUnDelValSetCmd() (*osmocli.TxCliDesc, *types.MsgUndelegateFromValidatorSet) {
// 	return &osmocli.TxCliDesc{
// 		Use:     "undelegate-valset",
// 		Short:   "UnDelegate tokens from existing valset using delegatorAddress and tokenAmount.",
// 		Example: "osmosisd tx valset-pref undelegate-valset osmo1... 100stake",
// 		NumArgs: 2,
// 	}, &types.MsgUndelegateFromValidatorSet{}
// }

func NewUndelRebalancedValSetCmd() (*osmocli.TxCliDesc, *types.MsgUndelegateFromRebalancedValidatorSet) {
	return &osmocli.TxCliDesc{
		Use:     "undelegate-rebalanced-valset",
		Short:   "Undelegate tokens from rebalanced valset using delegatorAddress and tokenAmount.",
		Long:    "Undelegates from an existing valset, but calculates the valset weights based on current user delegations.",
		Example: "osmosisd tx valset-pref undelegate-rebalanced-valset osmo1... 100stake",
		NumArgs: 2,
	}, &types.MsgUndelegateFromRebalancedValidatorSet{}
}

func NewReDelValSetCmd() (*osmocli.TxCliDesc, *types.MsgRedelegateValidatorSet) {
	return &osmocli.TxCliDesc{
		Use:              "redelegate-valset",
		Short:            "Redelegate tokens from existing validators to new sets of validators",
		Example:          "osmosisd tx valset-pref redelegate-valset  osmo1... osmovaloper1efg...,osmovaloper1hij...  0.56,0.44",
		NumArgs:          3,
		ParseAndBuildMsg: NewMsgReDelValidatorSetPreference,
	}, &types.MsgRedelegateValidatorSet{}
}

func NewWithRewValSetCmd() (*osmocli.TxCliDesc, *types.MsgWithdrawDelegationRewards) {
	return &osmocli.TxCliDesc{
		Use:     "withdraw-reward-valset",
		Short:   "Withdraw delegation reward form the new validator set.",
		Example: "osmosisd tx valset-pref withdraw-valset osmo1...",
		NumArgs: 1,
	}, &types.MsgWithdrawDelegationRewards{}
}

func NewMsgSetValidatorSetPreference(clientCtx client.Context, args []string, fs *pflag.FlagSet) (sdk.Msg, error) {
	delAddr, err := sdk.AccAddressFromBech32(args[0])
	if err != nil {
		return nil, err
	}

	valset, err := ValidateValAddrAndWeight(args)
	if err != nil {
		return nil, err
	}

	return types.NewMsgSetValidatorSetPreference(
		delAddr,
		valset,
	), nil
}

func NewMsgReDelValidatorSetPreference(clientCtx client.Context, args []string, fs *pflag.FlagSet) (sdk.Msg, error) {
	delAddr, err := sdk.AccAddressFromBech32(args[0])
	if err != nil {
		return nil, err
	}

	valset, err := ValidateValAddrAndWeight(args)
	if err != nil {
		return nil, err
	}

	return types.NewMsgRedelegateValidatorSet(
		delAddr,
		valset,
	), nil
}

func ValidateValAddrAndWeight(args []string) ([]types.ValidatorPreference, error) {
	var valAddrs []string
	valAddrs = append(valAddrs, strings.Split(args[1], ",")...)

	weights, err := osmoutils.ParseSdkDecFromString(args[2], ",")
	if err != nil {
		return nil, err
	}

	if len(valAddrs) != len(weights) {
		return nil, errors.New("the length of validator addresses and weights not matched")
	}

	if len(valAddrs) == 0 {
		return nil, errors.New("records is empty")
	}

	var valset []types.ValidatorPreference
	for i, val := range valAddrs {
		valset = append(valset, types.ValidatorPreference{
			ValOperAddress: val,
			Weight:         weights[i],
		})
	}

	return valset, nil
}
