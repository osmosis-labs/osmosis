package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/spf13/cobra"

	"github.com/c-osmosis/osmosis/app/params"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
)

const (
	flagSnapshotOutput = "snapshot-output"
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
	AtomAddress string  `json:"atom_address"`
	AtomBalance sdk.Int `json:"atom_balance"`
	AtomPercent sdk.Dec `json:"atom_ownership_percentage"`
	OsmoAddress string  `json:"osmo_address"`
	OsmoBalance sdk.Int `json:"osmo_balance"`
	OsmoPercent sdk.Dec `json:"osmo_ownership_percentage"`
}

// setCosmosBech32Prefixes set config for cosmos address system
func setCosmosBech32Prefixes() {
	defaultConfig := sdk.NewConfig()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(defaultConfig.GetBech32AccountAddrPrefix(), defaultConfig.GetBech32AccountPubPrefix())
	config.SetBech32PrefixForValidator(defaultConfig.GetBech32ValidatorAddrPrefix(), defaultConfig.GetBech32ValidatorPubPrefix())
	config.SetBech32PrefixForConsensusNode(defaultConfig.GetBech32ConsensusAddrPrefix(), defaultConfig.GetBech32ConsensusPubPrefix())
}

// ExportAirdropFromGenesisCmd returns add-genesis-account cobra Command.
func ExportAirdropFromGenesisCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-airdrop-genesis [denom] [file] [totalAmount]",
		Short: "Import balances from provided genesis to {FlagHome}/genesis.json",
		Long: `Import balances from provided genesis to {FlagHome}/genesis.json
Download:
	https://raw.githubusercontent.com/cephalopodequipment/cosmoshub-3/master/genesis.json
Init genesis file:
	osmosisd init mynode
Example:
	osmosisd export-airdrop-genesis uatom ../genesis.json 100000000000000 --snapshot-output="../snapshot.json"
	- Check genesis:
		file is at ~/.osmosisd/config/genesis.json
	- Snapshot
		file is at "../snapshot.json"
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			denom := args[0]
			filepath := args[1]
			osdenom := "uosmo"
			snapshotOutput, err := cmd.Flags().GetString(flagSnapshotOutput)
			if err != nil {
				return fmt.Errorf("failed to get snapshot directory: %w", err)
			}

			totalAmount, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("failed to parse totalAmount: %s", args[2])
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			jsonFile, err := os.Open(filepath)
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			byteValue, _ := ioutil.ReadAll(jsonFile)

			var genStateV036 GenesisStateV036

			setCosmosBech32Prefixes()
			err = aminoCodec.UnmarshalJSON(byteValue, &genStateV036)
			if err != nil {
				return err
			}

			snapshot := []SnapshotFields{}
			balanceIndexByAddress := make(map[string]int)
			totalAtomBalance := sdk.NewInt(0)
			for index, account := range genStateV036.AppState.Accounts {
				totalAtomBalance = totalAtomBalance.Add(account.Coins.AmountOf(denom))
				balanceIndexByAddress[account.Address.String()] = index
				snapshot = append(snapshot, SnapshotFields{
					AtomAddress: account.Address.String(),
					AtomBalance: account.Coins.AmountOf(denom),
					AtomPercent: sdk.NewDec(0),
				})
			}

			for _, delegation := range genStateV036.AppState.Staking.Delegations {
				address := delegation.DelegatorAddress
				index, ok := balanceIndexByAddress[address.String()]
				if !ok {
					continue
				}
				sharesInt := delegation.Shares.RoundInt()
				snapshot[index].AtomBalance = snapshot[index].AtomBalance.Add(sharesInt)
				totalAtomBalance = totalAtomBalance.Add(sharesInt)
			}

			for index, asnapshot := range snapshot {
				amt := asnapshot.AtomBalance
				percent := big.NewInt(0).Div(amt.Mul(sdk.NewInt(1000000)).BigInt(), totalAtomBalance.BigInt())
				snapshot[index].AtomPercent = sdk.NewDecFromBigIntWithPrec(percent, 4)
			}

			params.SetBech32Prefixes()

			balances := []banktypes.Balance{}
			for index, account := range genStateV036.AppState.Accounts {
				// fmt.Println("Address: " + account.Address.String())
				// fmt.Println("Amount: " + account.Coins.String())

				// create concrete account type based on input parameters
				var genAccount authtypes.GenesisAccount
				baseAccount := authtypes.NewBaseAccount(account.Address, nil, 0, 0)
				genAccount = baseAccount

				if err := genAccount.Validate(); err != nil {
					return fmt.Errorf("failed to validate new genesis account: %w", err)
				}

				// Add the new account to the set of genesis accounts and sanitize the
				// accounts afterwards.
				accs = append(accs, genAccount)
				accs = authtypes.SanitizeGenesisAccounts(accs)

				atomAmt := account.Coins.AmountOf(denom)
				osmoAmt, err := atomAmt.ToDec().ApproxSqrt()
				if err != nil {
					fmt.Println("failed to root atom balance", err)
					continue
				}
				coins := sdk.NewCoins(sdk.NewCoin(osdenom, osmoAmt.RoundInt()))
				address := account.Address
				balances = append(balances, banktypes.Balance{Address: address.String(), Coins: coins})
				balanceIndexByAddress[address.String()] = index
			}

			for _, delegation := range genStateV036.AppState.Staking.Delegations {
				address := delegation.DelegatorAddress
				shares := delegation.Shares
				index, ok := balanceIndexByAddress[address.String()]
				if !ok {
					continue
				}
				originAmt := sdk.NewInt(0)
				if len(balances[index].Coins) > 0 {
					originAmt = balances[index].Coins.AmountOf(osdenom)
				}
				osmoShareBonusRaw, err := shares.ApproxSqrt()
				if err != nil {
					fmt.Println("failed to root atom shares", err)
					continue
				}
				// apply 1.5x multiplier for
				osmoShareBonus := osmoShareBonusRaw.Mul(sdk.NewDecWithPrec(15, 10)).RoundInt()
				osmoAmt := originAmt.Add(osmoShareBonus)
				balances[index].Coins = sdk.NewCoins(sdk.NewCoin(osdenom, osmoAmt))
			}

			// normalize for total number of tokens to drop
			totalRaw := sdk.NewInt(0)
			for _, balance := range balances {
				totalRaw = totalRaw.Add(balance.Coins.AmountOf(osdenom))
			}
			for index, balance := range balances {
				amt := balance.Coins.AmountOf(osdenom)
				percent := big.NewInt(0).Div(amt.Mul(sdk.NewInt(1000000)).BigInt(), totalRaw.BigInt())
				snapshot[index].OsmoAddress = balance.Address
				snapshot[index].OsmoBalance = amt
				snapshot[index].OsmoPercent = sdk.NewDecFromBigIntWithPrec(percent, 4)
			}
			for i, balance := range balances {
				osmoAmtBI := balance.Coins.AmountOf(osdenom).BigInt()
				osmoAmtMulBI := osmoAmtBI.Mul(osmoAmtBI, totalAmount.BigInt())
				osmoAmtNormalBI := osmoAmtMulBI.Div(osmoAmtMulBI, totalRaw.BigInt())
				osmoAmtNormal := sdk.NewIntFromBigInt(osmoAmtNormalBI)
				balances[i].Coins = sdk.NewCoins(sdk.NewCoin(osdenom, osmoAmtNormal))
			}

			// remove empty accounts
			finalBalances := []banktypes.Balance{}
			totalDistr := sdk.NewInt(0)
			for _, balance := range balances {
				if balance.Coins.Empty() {
					continue
				}
				if balance.Coins.AmountOf(osdenom).Equal(sdk.NewInt(0)) {
					continue
				}
				finalBalances = append(finalBalances, balance)
				totalDistr = totalDistr.Add(balance.Coins.AmountOf(osdenom))
			}
			fmt.Println("total distributed amount:", totalDistr.String())
			fmt.Printf("cosmos accounts: %d\n", len(balances))
			fmt.Printf("empty drops: %d\n", len(balances)-len(finalBalances))
			fmt.Printf("available accounts: %d\n", len(finalBalances))

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(balances)

			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			err = genutil.ExportGenesisFile(genDoc, genFile)
			if err != nil {
				return err
			}

			// export snapshot directory
			snapshotJSON, err := aminoCodec.MarshalJSON(snapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}
			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flagSnapshotOutput, "", "Snapshot export file")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
