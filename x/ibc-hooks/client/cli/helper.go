package cli

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func getBech32CustomPrefix(config *sdk.Config, customPrefix string) (string, error) {
	if customPrefix == "" {
		return "", errors.New("trying to get bech32 prefix without specifying prefix name")
	}

	var prefix string
	switch customPrefix {
	case "account_addr":
		prefix = config.GetBech32AccountAddrPrefix()
	case "validator_addr":
		prefix = config.GetBech32ValidatorAddrPrefix()
	case "consensus_addr":
		prefix = config.GetBech32ConsensusAddrPrefix()
	case "account_pub":
		prefix = config.GetBech32AccountPubPrefix()
	case "validator_pub":
		prefix = config.GetBech32ValidatorPubPrefix()
	case "consensus_pub":
		prefix = config.GetBech32ConsensusPubPrefix()
	default:
		return "", fmt.Errorf("Prefix %s is invalid", customPrefix)
	}
	return prefix, nil
}
