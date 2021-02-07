package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// GenesisStateV036 is minimum structure to import airdrop accounts
type GenesisStateV036 struct {
	AppState AppStateV036 `json:"app_state"`
}

// AppStateV036 is app state structure for app state
type AppStateV036 struct {
	Accounts []v036genaccounts.GenesisAccount `json:"accounts"`
	Staking  v036staking.GenesisState         `json:"staking"`
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
	osmosisd export-airdrop-genesis uatom ../genesis.json 100000000000000
Check genesis:
  file is at ~/.osmosisd/config/genesis.json
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
			params.SetBech32Prefixes()

			balanceIndexByAddress := make(map[string]int)
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
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
