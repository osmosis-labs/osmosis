package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

type DeriveSnapshot struct {
	NumberAccounts uint64                    `json:"num_accounts"`
	Accounts       map[string]DerivedAccount `json:"accounts"`
}

// DerivedAccount provide fields of snapshot per account
type DerivedAccount struct {
	Address  string    `json:"address"`
	Balances sdk.Coins `json:"balance"`
	Staked   sdk.Int   `json:"staked"`
	Unstaked sdk.Int   `json:"unstaked"`
	Bonded   sdk.Coins `json:"bonded"`
}

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json
func ExportDeriveBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-derive-balances [input-genesis-file] [output-snapshot-json]",
		Short: "Export a derive balances from a provided genesis export",
		Long: `Export a derive balances from a provided genesis export
Example:
	osmosisd export-derive-balances ../genesis.json ../snapshot.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genesisFile := args[0]
			snapshotOutput := args[1]

			// Read genesis file
			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var doc tmtypes.GenesisDoc
			err = tmjson.Unmarshal(byteValue, &doc)
			if err != nil {
				return err
			}

			genState := make(map[string]json.RawMessage)
			err = json.Unmarshal(doc.AppState, &genState)
			if err != nil {
				panic(err)
			}

			authGenesis := authtypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["auth"], &authGenesis)
			accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
			if err != nil {
				panic(err)
			}
			accounts = authtypes.SanitizeGenesisAccounts(accounts)

			// Produce the map of address to total atom balance, both staked and unstaked
			snapshotAccs := make(map[string]DerivedAccount)
			for _, account := range accounts {

				snapshotAccs[account.GetAddress().String()] = DerivedAccount{
					Address:  account.GetAddress().String(),
					Balances: sdk.Coins{},
					Staked:   sdk.ZeroInt(),
					Bonded:   sdk.Coins{},
				}
			}

			bankGenesis := banktypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["bank"], &bankGenesis)
			for _, balance := range bankGenesis.Balances {
				address := balance.Address
				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for bank balance")
				}

				acc.Balances = balance.Coins
				snapshotAccs[address] = acc
			}

			stakingGenesis := stakingtypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["staking"], &stakingGenesis)
			for _, unbonding := range stakingGenesis.UnbondingDelegations {
				address := unbonding.DelegatorAddress
				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for unbonding")
				}

				unbondingOsmos := sdk.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingOsmos = unbondingOsmos.Add(entry.Balance)
				}

				acc.Unstaked = acc.Unstaked.Add(unbondingOsmos)

				snapshotAccs[address] = acc
			}

			// Make a map from validator operator address to the v036 validator type
			validators := make(map[string]stakingtypes.Validator)
			for _, validator := range stakingGenesis.Validators {
				validators[validator.OperatorAddress] = validator
			}

			for _, delegation := range stakingGenesis.Delegations {
				address := delegation.DelegatorAddress

				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for delegation")
				}

				val := validators[delegation.ValidatorAddress]
				stakedOsmos := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.Staked = acc.Staked.Add(stakedOsmos)

				snapshotAccs[address] = acc
			}

			lockupGenesis := lockuptypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["lockup"], &lockupGenesis)
			for _, lock := range lockupGenesis.Locks {
				address := lock.Owner

				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for lock")
				}

				acc.Bonded = acc.Bonded.Add(lock.Coins...)
				snapshotAccs[address] = acc
			}

			snapshot := DeriveSnapshot{
				NumberAccounts: uint64(len(snapshotAccs)),
				Accounts:       snapshotAccs,
			}

			gammGenesis := gammtypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["gamm"], &gammGenesis)

			fmt.Printf("# accounts: %d\n", len(snapshotAccs))

			// export snapshot json
			snapshotJSON, err := json.MarshalIndent(snapshot, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}

			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
