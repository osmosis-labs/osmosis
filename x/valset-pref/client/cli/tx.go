package valsetprefcli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

func GetTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	txCmd.AddCommand(
		NewSetValSetCmd(),
		NewDelValSetCmd(),
		NewUnDelValSetCmd(),
	)

	return txCmd
}

func NewSetValSetCmd() *cobra.Command {
	return osmocli.TxCliDesc{
		Use:              "set-valset [delegator_addr] [validators] [weights]",
		Short:            "Creates a new validator set for the delegator with valOperAddress and weight",
		Example:          "osmosisd tx validatorsetpreference set-valset osmo1... osmovaloper1abc...,osmovaloper1def...  0.56,0.44",
		NumArgs:          3,
		ParseAndBuildMsg: NewMsgSetValidatorSetPreference,
	}.BuildCommandCustomFn()
}

func NewDelValSetCmd() *cobra.Command {
	return osmocli.TxCliDesc{
		Use:              "delegate-valset [delegator_addr] [amount]",
		Short:            "Delegate tokens to existing valset using delegatorAddress and tokenAmount.",
		Example:          "osmosisd tx validatorsetpreference delegate-valset  osmo1... 100stake",
		NumArgs:          2,
		ParseAndBuildMsg: NewMsgDelegateToValidatorSet,
	}.BuildCommandCustomFn()
}

func NewUnDelValSetCmd() *cobra.Command {
	return osmocli.TxCliDesc{
		Use:              "undelegate-valset [delegator_addr] [amount]",
		Short:            "UnDelegate tokens from existing valset using delegatorAddress and tokenAmount.",
		Example:          "osmosisd tx validatorsetpreference undelegate-valset  osmo1... 100stake",
		NumArgs:          2,
		ParseAndBuildMsg: NewMsgUndelegateFromValidatorSet,
	}.BuildCommandCustomFn()
}

func NewMsgSetValidatorSetPreference(clientCtx client.Context, args []string, fs *pflag.FlagSet) (sdk.Msg, error) {
	delAddr, err := sdk.AccAddressFromBech32(args[0])
	if err != nil {
		return nil, err
	}

	valAddrs, err := osmoutils.ParseSdkValAddressFromString(args[1], ",")
	if err != nil {
		return nil, err
	}

	weights, err := osmoutils.ParseSdkDecFromString(args[2], ",")
	if err != nil {
		return nil, err
	}

	if len(valAddrs) != len(weights) {
		return nil, fmt.Errorf("the length of validator addresses and weights not matched")
	}

	if len(valAddrs) == 0 {
		return nil, fmt.Errorf("records is empty")
	}

	var valset []types.ValidatorPreference
	for i, val := range valAddrs {
		valset = append(valset, types.ValidatorPreference{
			Weight:         weights[i],
			ValOperAddress: val.String(),
		})
	}

	return types.NewMsgSetValidatorSetPreference(
		delAddr,
		valset,
	), nil
}

func NewMsgDelegateToValidatorSet(clientCtx client.Context, args []string, fs *pflag.FlagSet) (sdk.Msg, error) {
	delAddr, err := sdk.AccAddressFromBech32(args[0])
	if err != nil {
		return nil, err
	}

	delegationAmount, err := sdk.ParseCoinNormalized(args[1])
	if err != nil {
		return nil, err
	}

	return types.NewMsgDelegateToValidatorSet(
		delAddr,
		delegationAmount,
	), nil
}

func NewMsgUndelegateFromValidatorSet(clientCtx client.Context, args []string, fs *pflag.FlagSet) (sdk.Msg, error) {
	delAddr, err := sdk.AccAddressFromBech32(args[0])
	if err != nil {
		return nil, err
	}

	undelegationAmount, err := sdk.ParseCoinNormalized(args[1])
	if err != nil {
		return nil, err
	}

	return types.NewMsgUndelegateFromValidatorSet(
		delAddr,
		undelegationAmount,
	), nil
}
