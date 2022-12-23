package valsetprefcli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

func GetTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewSetValSetCmd)
	osmocli.AddTxCmd(txCmd, NewDelValSetCmd)
	osmocli.AddTxCmd(txCmd, NewUnDelValSetCmd)
	return txCmd
}

func NewSetValSetCmd() (*osmocli.TxCliDesc, *types.MsgSetValidatorSetPreference) {
	return &osmocli.TxCliDesc{
		Use:     "set-valset [delegator_addr] [validators] [weights]",
		Short:   "Creates a new validator set for the delegator with valOperAddress and weight",
		Example: "osmosisd tx validatorsetpreference set-valset osmo1... osmovaloper1abc...,osmovaloper1def...  0.56,0.44",
	}, &types.MsgSetValidatorSetPreference{}
}

func NewDelValSetCmd() (*osmocli.TxCliDesc, *types.MsgDelegateToValidatorSet) {
	return &osmocli.TxCliDesc{
		Use:     "delegate-valset [delegator_addr] [amount]",
		Short:   "Delegate tokens to existing valset using delegatorAddress and tokenAmount.",
		Example: "osmosisd tx validatorsetpreference delegate-valset  osmo1... 100stake",
	}, &types.MsgDelegateToValidatorSet{}
}

func NewUnDelValSetCmd() (*osmocli.TxCliDesc, *types.MsgUndelegateFromValidatorSet) {
	return &osmocli.TxCliDesc{
		Use:     "undelegate-valset [delegator_addr] [amount]",
		Short:   "UnDelegate tokens from existing valset using delegatorAddress and tokenAmount.",
		Example: "osmosisd tx validatorsetpreference undelegate-valset  osmo1... 100stake",
	}, &types.MsgUndelegateFromValidatorSet{}
}
