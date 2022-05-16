package testnetify

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v8/app"
)

var (
	flagTestnetParams = "empty"
	valConsBech32     = "osmovalcons"
)

// TODO: Add params for min num validators for consensus.
type TestnetParams struct {
	ValidatorConsensusPubkeys []string
	ValidatorOperatorPubkeys  []string
	OutputExportFilepath      string
}

type ValidatorDetails struct {
	// e.g. 16A169951A878247DBE258FDDC71638F6606D156
	// Only appears once
	ConsAddressHex string
	// e.g. b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=
	ConsPubkeyRaw string
	// e.g. osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n
	OperatorAddress string
}

//nolint:ineffassign
func defaultTestnetParams() TestnetParams {
	/* priv_validator_key.json file
	{
		"address": "8B78E478777427CC3906B8234CB72BCEA2C78E83",
		"pub_key": {
			"type": "tendermint/PubKeyEd25519",
			"value": "2OpBuqaXvXQ+lSxAoT1S7Jfyr56KiakTzvuFiuJK+X4="
		},
		"priv_key": {
			"type": "tendermint/PrivKeyEd25519",
			"value": "3OLLoEfdT+ZrLqpRCvytpXrhgKfeEBBoeaoXe1p3/mjY6kG6ppe9dD6VLEChPVLsl/KvnoqJqRPO+4WK4kr5fg=="
		}
	}
	*/
	consensusPubkey := "2OpBuqaXvXQ+lSxAoT1S7Jfyr56KiakTzvuFiuJK+X4="
	// mnemonic kitchen comic flower drip sick prize account cheese truth income weekend nominee segment punch call satisfy captain earth ethics wasp clump tunnel orchard advance
	operatorPubkey := "osmopub1addwnpepq0wv95fxk32z90u42t3df2l5pdtngvg20rkalv5vt2d7n5q5ekk35d8hh20"
	operatorValOper := "osmovaloper1qye772qje88p7ggtzrvl9nxvty6dkuusvpqhac"
	operatorValOper += ""
	return TestnetParams{
		ValidatorConsensusPubkeys: []string{consensusPubkey},
		ValidatorOperatorPubkeys:  []string{operatorPubkey},
		OutputExportFilepath:      "new_testnet_genesis.json",
	}
}

func (params TestnetParams) NewValidatorDetails() []ValidatorDetails {
	numValidators := len(params.ValidatorConsensusPubkeys)
	outputDeets := make([]ValidatorDetails, 0, numValidators)
	for i := 0; i < numValidators; i++ {
		deets := ValidatorDetails{
			ConsPubkeyRaw:   params.ValidatorConsensusPubkeys[i],
			OperatorAddress: params.ValidatorConsensusPubkeys[i],
		}
		outputDeets = append(outputDeets, deets)
	}
	return outputDeets
}

func replaceValidatorDetails(genesis app.GenesisState, params TestnetParams) {
	oldValidatorDetails := getValidatorDetailsToReplace(genesis, params)
	newValidatorDetails := params.NewValidatorDetails()
	numValidators := len(oldValidatorDetails)
	for i := 0; i < numValidators; i++ {
		oldDeets := oldValidatorDetails[i]
		newDeets := newValidatorDetails[i]

		// TODO: Use more module-focused find replaces
		replaceConsAddrHex(genesis, oldDeets.ConsAddressHex, newDeets.ConsAddressHex)
		replaceAllInGenesis(genesis, oldDeets.ConsPubkeyRaw, newDeets.ConsPubkeyRaw)
		replaceAllInGenesis(genesis, oldDeets.OperatorAddress, newDeets.OperatorAddress)
	}
}

func getValidatorDetailsToReplace(genesis app.GenesisState, params TestnetParams) []ValidatorDetails {
	// TODO: Don't hardcode to sentinel, instead all validators from chain
	// and get top N validators
	// consAddr only has 1 instance
	validatorConsAddr := "16A169951A878247DBE258FDDC71638F6606D156"
	// Only has 2 instances
	validatorConsPubkey := "b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU="
	// Sometimes appears in slashing
	// validatorConsBech32Addr := "osmosvalcons1z6skn9g6s7py0klztr7acutr3anqd52kuhdukh"
	validatorOperBech32Addr := "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n"

	SentinelValidatorDetails := ValidatorDetails{
		ConsAddressHex:  validatorConsAddr,
		ConsPubkeyRaw:   validatorConsPubkey,
		OperatorAddress: validatorOperBech32Addr,
	}
	return []ValidatorDetails{
		SentinelValidatorDetails,
	}
}

func updateChainId(genesis app.GenesisState) {
	// TODO: Implement
}

func clearIBC(genesis app.GenesisState) {
	// TODO: Implement
}

func loadTestnetParams(cmd *cobra.Command) (TestnetParams, error) {
	testnetParamPath, err := cmd.Flags().GetString(flagTestnetParams)
	if err != nil {
		return TestnetParams{}, err
	}
	if testnetParamPath == "empty" {
		return defaultTestnetParams(), nil
	}
	panic("TODO: Go read testnet params from a file")
}
