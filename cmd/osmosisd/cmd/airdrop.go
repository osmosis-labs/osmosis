package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
)

const (
	flagOsmoSupply = "osmo-supply"
)

// GenesisStateV036 is minimum structure to import airdrop accounts
type GenesisStateV036 struct {
	AppState AppStateV036 `json:"app_state"`
}

// AppStateV036 is app state structure for app state
type AppStateV036 struct {
	Accounts []v036genaccounts.GenesisAccount `json:"accounts"`
	Staking  v036staking.GenesisState         `json:"staking"`
}

// SnapshotFields provide fields of snapshot per account
type SnapshotFields struct {
	AtomAddress           string  `json:"atom_address"`
	AtomBalance           sdk.Int `json:"atom_balance"`
	AtomStakedBalance     sdk.Int `json:"atom_staked_balance"`
	AtomUnstakedBalance   sdk.Int `json:"atom_unstaked_balance"`
	AtomStakedPercent     sdk.Dec `json:"atom_staked_percent"`
	AtomOwnershipPercent  sdk.Dec `json:"atom_ownership_percent"`
	OsmoNormalizedBalance sdk.Int `json:"osmo_balance_normalized"`
	OsmoBalance           sdk.Int `json:"osmo_balance"`
	OsmoBalanceBonus      sdk.Int `json:"osmo_balance_bonus"`
	OsmoBalanceBase       sdk.Int `json:"osmo_balance_base"`
	OsmoPercent           sdk.Dec `json:"osmo_ownership_percent"`
}

// setCosmosBech32Prefixes set config for cosmos address system
func setCosmosBech32Prefixes() {
	defaultConfig := sdk.NewConfig()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(defaultConfig.GetBech32AccountAddrPrefix(), defaultConfig.GetBech32AccountPubPrefix())
	config.SetBech32PrefixForValidator(defaultConfig.GetBech32ValidatorAddrPrefix(), defaultConfig.GetBech32ValidatorPubPrefix())
	config.SetBech32PrefixForConsensusNode(defaultConfig.GetBech32ConsensusAddrPrefix(), defaultConfig.GetBech32ConsensusPubPrefix())
}

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided cosmos-sdk v0.36 genesis export.
func ExportAirdropSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-airdrop-snapshot [airdrop-to-denom] [input-genesis-file] [output-snapshot-json] --osmo-supply=[osmos-genesis-supply]",
		Short: "Export a quadratic fairdrop snapshot from a provided cosmos-sdk v0.36 genesis export",
		Long: `Export a quadratic fairdrop snapshot from a provided cosmos-sdk v0.36 genesis export
Sample genesis file:
	https://raw.githubusercontent.com/cephalopodequipment/cosmoshub-3/master/genesis.json
Example:
	osmosisd export-airdrop-genesis uatom ~/.gaiad/config/genesis.json ../snapshot.json --osmo-supply=100000000000000
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
	- Snapshot
		file is at "../snapshot.json"
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			denom := args[0]
			genesisFile := args[1]
			snapshotOutput := args[2]

			osmoSupplyStr, err := cmd.Flags().GetString(flagOsmoSupply)
			if err != nil {
				return fmt.Errorf("failed to get osmo total supply: %w", err)
			}
			osmoSupply, ok := sdk.NewIntFromString(osmoSupplyStr)
			if !ok {
				return fmt.Errorf("failed to parse osmo supply: %s", osmoSupplyStr)
			}

			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var genStateV036 GenesisStateV036

			setCosmosBech32Prefixes()
			err = aminoCodec.UnmarshalJSON(byteValue, &genStateV036)
			if err != nil {
				return err
			}

			snapshot := make(map[string]SnapshotFields)

			totalAtomBalance := sdk.NewInt(0)
			for _, account := range genStateV036.AppState.Accounts {

				balance := account.Coins.AmountOf(denom)
				totalAtomBalance = totalAtomBalance.Add(balance)

				if account.ModuleName != "" {
					continue
				}

				snapshot[account.Address.String()] = SnapshotFields{
					AtomAddress:         account.Address.String(),
					AtomBalance:         balance,
					AtomUnstakedBalance: balance,
					AtomStakedBalance:   sdk.ZeroInt(),
				}
			}

			for _, unbonding := range genStateV036.AppState.Staking.UnbondingDelegations {
				address := unbonding.DelegatorAddress.String()
				acc, ok := snapshot[address]
				if !ok {
					panic("no account found for unbonding")
				}

				unbondingAtoms := sdk.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingAtoms = unbondingAtoms.Add(entry.Balance)
				}

				acc.AtomBalance = acc.AtomBalance.Add(unbondingAtoms)
				acc.AtomUnstakedBalance = acc.AtomUnstakedBalance.Add(unbondingAtoms)

				snapshot[address] = acc
			}

			validators := make(map[string]v036staking.Validator)
			for _, validator := range genStateV036.AppState.Staking.Validators {
				validators[validator.OperatorAddress.String()] = validator
			}

			for _, delegation := range genStateV036.AppState.Staking.Delegations {
				address := delegation.DelegatorAddress.String()

				acc, ok := snapshot[address]
				if !ok {
					panic("no account found for delegation")
				}

				val := validators[delegation.ValidatorAddress.String()]
				stakedAtoms := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.AtomBalance = acc.AtomBalance.Add(stakedAtoms)
				acc.AtomStakedBalance = acc.AtomStakedBalance.Add(stakedAtoms)

				snapshot[address] = acc
			}

			totalOsmoBalance := sdk.NewInt(0)

			onePointFive := sdk.MustNewDecFromStr("1.5")

			for address, acc := range snapshot {
				allAtoms := acc.AtomBalance.ToDec()

				acc.AtomOwnershipPercent = allAtoms.QuoInt(totalAtomBalance)

				if allAtoms.IsZero() {
					acc.AtomStakedPercent = sdk.ZeroDec()
					acc.OsmoBalanceBase = sdk.ZeroInt()
					acc.OsmoBalanceBonus = sdk.ZeroInt()
					acc.OsmoBalance = sdk.ZeroInt()
					snapshot[address] = acc
					continue
				}

				stakedAtoms := acc.AtomStakedBalance.ToDec()
				stakedPercent := stakedAtoms.Quo(allAtoms)
				acc.AtomStakedPercent = stakedPercent

				baseOsmo, err := allAtoms.ApproxSqrt()
				if err != nil {
					panic(fmt.Sprintf("failed to root atom balance: %s", err))
				}
				acc.OsmoBalanceBase = baseOsmo.RoundInt()

				bonusOsmo := baseOsmo.Mul(onePointFive).Mul(stakedPercent)
				acc.OsmoBalanceBonus = bonusOsmo.RoundInt()

				allOsmo := baseOsmo.Add(bonusOsmo)
				acc.OsmoBalance = allOsmo.RoundInt()

				totalOsmoBalance = totalOsmoBalance.Add(allOsmo.RoundInt())

				if allAtoms.LTE(sdk.NewDec(1000000)) {
					acc.OsmoBalanceBase = sdk.ZeroInt()
					acc.OsmoBalanceBonus = sdk.ZeroInt()
					acc.OsmoBalance = sdk.ZeroInt()
				}

				snapshot[address] = acc
			}

			// normalize to desired genesis osmo supply
			noarmalizationFactor := osmoSupply.ToDec().Quo(totalOsmoBalance.ToDec())

			for address, acc := range snapshot {
				acc.OsmoPercent = acc.OsmoBalance.ToDec().Quo(totalOsmoBalance.ToDec())

				acc.OsmoNormalizedBalance = acc.OsmoBalance.ToDec().Mul(noarmalizationFactor).RoundInt()

				snapshot[address] = acc
			}

			fmt.Printf("cosmos accounts: %d\n", len(snapshot))
			fmt.Printf("atomTotalSupply: %s\n", totalAtomBalance.String())
			fmt.Printf("osmoTotalSupply (pre-normalization): %s\n", totalOsmoBalance.String())

			// export snapshot json
			snapshotJSON, err := aminoCodec.MarshalJSON(snapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}
			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	cmd.Flags().String(flagOsmoSupply, "", "OSMO total genesis supply")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
