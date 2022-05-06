package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

				// attempt to lookup address from Keybase if no address was provided
				kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf)
				if err != nil {
					return err
				}

				info, err := kb.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keybase: %w", err)
				}

				addr = info.GetAddress()
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			vestingStart, _ := cmd.Flags().GetInt64(flagVestingStart)
			vestingEnd, _ := cmd.Flags().GetInt64(flagVestingEnd)
			vestingAmtStr, _ := cmd.Flags().GetString(flagVestingAmt)

			vestingAmt, err := sdk.ParseCoinsNormalized(vestingAmtStr)
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount: %w", err)
			}

			// create concrete account type based on input parameters
			var genAccount authtypes.GenesisAccount

			balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
			baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)

			if !vestingAmt.IsZero() {
				baseVestingAccount := authvesting.NewBaseVestingAccount(baseAccount, vestingAmt.Sort(), vestingEnd)

				if (balances.Coins.IsZero() && !baseVestingAccount.OriginalVesting.IsZero()) ||
					baseVestingAccount.OriginalVesting.IsAnyGT(balances.Coins) {
					return errors.New("vesting amount cannot be greater than total amount")
				}

				switch {
				case vestingStart != 0 && vestingEnd != 0:
					genAccount = authvesting.NewContinuousVestingAccountRaw(baseVestingAccount, vestingStart)

				case vestingEnd != 0:
					genAccount = authvesting.NewDelayedVestingAccountRaw(baseVestingAccount)

				default:
					return errors.New("invalid vesting parameters; must supply start and end time or end time")
				}
			} else {
				genAccount = baseAccount
			}

			if err := genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
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

			if accs.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
			}

			// Add the new account to the set of genesis accounts and sanitize the
			// accounts afterwards.
			accs = append(accs, genAccount)
			accs = authtypes.SanitizeGenesisAccounts(accs)

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
			bankGenState.Balances = append(bankGenState.Balances, balances)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

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
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Int64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// func GetOsmoSnapshot(inputFile string) (Snapshot, error) {
// 	snapshotJSON, err := os.Open(inputFile)
// 	if err != nil {
// 		return Snapshot{}, err
// 	}
// 	defer snapshotJSON.Close()

// 	byteValue, err := ioutil.ReadAll(snapshotJSON)
// 	if err != nil {
// 		return Snapshot{}, err
// 	}

// 	snapshot := Snapshot{}
// 	err = json.Unmarshal(byteValue, &snapshot)
// 	if err != nil {
// 		return Snapshot{}, err
// 	}
// 	return snapshot, nil
// }

// func GetIonSnapshot(inputFile string) (map[string]int64, error) {
// 	ionJSON, err := os.Open(inputFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer ionJSON.Close()
// 	byteValue2, _ := ioutil.ReadAll(ionJSON)

// 	var ionAmts map[string]int64
// 	err = json.Unmarshal(byteValue2, &ionAmts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ionAmts, nil
// }

// func CosmosToOsmoAddress(cosmosAddr string) (string, error) {
// 	_, bz, err := bech32.DecodeAndConvert(cosmosAddr)
// 	if err != nil {
// 		return "", err
// 	}

// 	convertedAddr, err := bech32.ConvertAndEncode("osmo", bz)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return convertedAddr, nil
// }

// func GetAirdropAccountsCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "get-airdrop-accounts [input-snapshot-file] [input-ions-file] [output-file]",
// 		Short: "Get list of all accounts that are being airdropped to at genesis",
// 		Long: `Get list of all accounts that are being airdropped to at genesis
// Both OSMO and ION recipients. If erroring, ensure to 'git lfs pull'

// Example:
// 	osmosisd import-genesis-accounts-from-snapshot networks/cosmoshub-3/snapshot.json networks/osmosis-1/ions.json output_address.json
// `,
// 		Args: cobra.ExactArgs(3),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			airdropAddrs := map[string]bool{}

// 			// Read snapshot file
// 			snapshot, err := GetOsmoSnapshot(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			// Read ions file
// 			ionAmts, err := GetIonSnapshot(args[1])
// 			if err != nil {
// 				return err
// 			}

// 			for _, acc := range snapshot.Accounts {
// 				if !acc.OsmoBalance.Equal(sdk.ZeroInt()) {
// 					osmoAddr, err := CosmosToOsmoAddress(acc.AtomAddress)
// 					if err != nil {
// 						return err
// 					}

// 					airdropAddrs[osmoAddr] = true
// 				}
// 			}

// 			for addr := range ionAmts {
// 				ionAddr, err := CosmosToOsmoAddress(addr)
// 				if err != nil {
// 					return err
// 				}

// 				airdropAddrs[ionAddr] = true
// 			}

// 			airdropAddrsJSON, err := json.Marshal(airdropAddrs)
// 			if err != nil {
// 				return err
// 			}
// 			err = ioutil.WriteFile(args[2], airdropAddrsJSON, 0644)
// 			return err
// 		},
// 	}

// 	flags.AddQueryFlagsToCmd(cmd)

// 	return cmd
// }

// func ImportGenesisAccountsFromSnapshotCmd(defaultNodeHome string) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "import-genesis-accounts-from-snapshot [input-snapshot-file] [input-ions-file]",
// 		Short: "Import genesis accounts from fairdrop snapshot.json and an ions.json",
// 		Long: `Import genesis accounts from fairdrop snapshot.json
// 20% of airdrop amount is liquid in accounts.
// The remaining is placed in the claims module.

// Must also pass in an ions.json file to airdrop genesis ions
// Example:
// 	osmosisd import-genesis-accounts-from-snapshot ../snapshot.json ../ions.json
// 	- Check input genesis:
// 		file is at ~/.osmosisd/config/genesis.json
// `,
// 		Args: cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx := client.GetClientContextFromCmd(cmd)
// 			depCdc := clientCtx.Codec
// 			cdc := depCdc.(codec.Codec)
// 			// aminoCodec := clientCtx.LegacyAmino.Amino

// 			serverCtx := server.GetServerContextFromCmd(cmd)
// 			config := serverCtx.Config

// 			config.SetRoot(clientCtx.HomeDir)

// 			genFile := config.GenesisFile()
// 			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
// 			if err != nil {
// 				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
// 			}

// 			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

// 			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
// 			if err != nil {
// 				return fmt.Errorf("failed to get accounts from any: %w", err)
// 			}

// 			// Read snapshot file
// 			snapshot, err := GetOsmoSnapshot(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			// Read ions file
// 			ionAmts, err := GetIonSnapshot(args[1])
// 			if err != nil {
// 				return err
// 			}

// 			// get genesis params
// 			genesisParams := MainnetGenesisParams()

// 			nonAirdropAccs := make(map[string]sdk.Coins)

// 			for _, acc := range genesisParams.StrategicReserveAccounts {
// 				nonAirdropAccs[acc.Address] = acc.GetCoins()
// 			}

// 			for _, acc := range genesisParams.MintParams.WeightedDeveloperRewardsReceivers {
// 				if _, ok := nonAirdropAccs[acc.Address]; !ok {
// 					nonAirdropAccs[acc.Address] = sdk.NewCoins()
// 				}

// 			}

// 			for addr, amt := range ionAmts {
// 				address, err := CosmosToOsmoAddress(addr)
// 				if err != nil {
// 					return err
// 				}

// 				if val, ok := nonAirdropAccs[address]; ok {
// 					nonAirdropAccs[address] = val.Add(sdk.NewCoin("uion", sdk.NewInt(amt).MulRaw(1_000_000)))
// 				} else {
// 					nonAirdropAccs[address] = sdk.NewCoins(sdk.NewCoin("uion", sdk.NewInt(amt).MulRaw(1_000_000)))
// 				}
// 			}

// 			// figure out normalizationFactor to normalize snapshot balances to desired airdrop supply
// 			normalizationFactor := genesisParams.AirdropSupply.ToDec().QuoInt(snapshot.TotalOsmosAirdropAmount)
// 			fmt.Printf("normalization factor: %s\n", normalizationFactor)

// 			bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)

// 			liquidBalances := bankGenState.Balances
// 			claimRecords := []claimtypes.ClaimRecord{}
// 			claimModuleAccountBalance := sdk.NewInt(0)

// 			// for each account in the snapshot
// 			for _, acc := range snapshot.Accounts {
// 				// convert cosmos address to osmo address
// 				address, err := CosmosToOsmoAddress(acc.AtomAddress)
// 				if err != nil {
// 					return err
// 				}

// 				// skip accounts with 0 balance
// 				if !acc.OsmoBalanceBase.IsPositive() {
// 					continue
// 				}

// 				// get normalized osmo balance for account
// 				normalizedOsmoBalance := acc.OsmoBalance.ToDec().Mul(normalizationFactor)

// 				// initial liquid amounts
// 				// We consistently round down to the nearest uosmo
// 				liquidAmount := normalizedOsmoBalance.Mul(sdk.MustNewDecFromStr("0.2")).TruncateInt() // 20% of airdrop amount
// 				liquidCoins := sdk.NewCoins(sdk.NewCoin(genesisParams.NativeCoinMetadatas[0].Base, liquidAmount))

// 				if coins, ok := nonAirdropAccs[address]; ok {
// 					liquidCoins = liquidCoins.Add(coins...)
// 					delete(nonAirdropAccs, address)
// 				}

// 				liquidBalances = append(liquidBalances, banktypes.Balance{
// 					Address: address,
// 					Coins:   liquidCoins,
// 				})

// 				// claimable balances
// 				claimableAmount := normalizedOsmoBalance.Mul(sdk.MustNewDecFromStr("0.8")).TruncateInt()

// 				claimRecords = append(claimRecords, claimtypes.ClaimRecord{
// 					Address:                address,
// 					InitialClaimableAmount: sdk.NewCoins(sdk.NewCoin(genesisParams.NativeCoinMetadatas[0].Base, claimableAmount)),
// 					ActionCompleted:        []bool{false, false, false, false},
// 				})

// 				claimModuleAccountBalance = claimModuleAccountBalance.Add(claimableAmount)

// 				// Add the new account to the set of genesis accounts
// 				sdkaddr, err := sdk.AccAddressFromBech32(address)
// 				if err != nil {
// 					return err
// 				}
// 				baseAccount := authtypes.NewBaseAccount(sdkaddr, nil, 0, 0)
// 				if err := baseAccount.Validate(); err != nil {
// 					return fmt.Errorf("failed to validate new genesis account: %w", err)
// 				}
// 				accs = append(accs, baseAccount)
// 			}

// 			// distribute remaining ions to accounts not in fairdrop
// 			for addr, remainingNonAirdrop := range nonAirdropAccs {
// 				// read address from snapshot
// 				address, err := sdk.AccAddressFromBech32(addr)
// 				if err != nil {
// 					return err
// 				}

// 				liquidBalances = append(liquidBalances, banktypes.Balance{
// 					Address: address.String(),
// 					Coins:   remainingNonAirdrop,
// 				})

// 				// Add the new account to the set of genesis accounts
// 				baseAccount := authtypes.NewBaseAccount(address, nil, 0, 0)
// 				if err := baseAccount.Validate(); err != nil {
// 					return fmt.Errorf("failed to validate new genesis account: %w", err)
// 				}
// 				accs = append(accs, baseAccount)
// 			}

// 			// auth module genesis
// 			accs = authtypes.SanitizeGenesisAccounts(accs)
// 			genAccs, err := authtypes.PackAccounts(accs)
// 			if err != nil {
// 				return fmt.Errorf("failed to convert accounts into any's: %w", err)
// 			}
// 			authGenState.Accounts = genAccs
// 			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
// 			}
// 			appState[authtypes.ModuleName] = authGenStateBz

// 			// bank module genesis
// 			bankGenState.Balances = banktypes.SanitizeGenesisBalances(liquidBalances)
// 			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
// 			}
// 			appState[banktypes.ModuleName] = bankGenStateBz

// 			// claim module genesis
// 			claimGenState := claimtypes.GetGenesisStateFromAppState(depCdc, appState)
// 			claimGenState.ModuleAccountBalance = sdk.NewCoin(genesisParams.NativeCoinMetadatas[0].Base, claimModuleAccountBalance)

// 			claimGenState.ClaimRecords = claimRecords
// 			claimGenStateBz, err := cdc.MarshalJSON(claimGenState)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal claim genesis state: %w", err)
// 			}
// 			appState[claimtypes.ModuleName] = claimGenStateBz

// 			// TODO: add remaining extra to community pool
// 			// The total airdrop osmo is a smidge short (~1 osmo) short of the stated 50M supply.
// 			// This is due to consistently rounding down.
// 			// We place this remaining 1 osmo into the community pool at genesis

// 			// sumAirdrop := sdk.Coins{}
// 			// for _, balance := range bankGenState.Balances {
// 			// 	sumAirdrop = sumAirdrop.Add(balance.Coins...)
// 			// }
// 			// for _, claim := range claimGenState.ClaimRecords {
// 			// 	sumAirdrop = sumAirdrop.Add(claim.InitialClaimableAmount...)
// 			// }

// 			// var distributionGenState distributiontypes.GenesisState

// 			// if appState[distributiontypes.ModuleName] != nil {
// 			// 	cdc.MustUnmarshalJSON(appState[distributiontypes.ModuleName], &distributionGenState)
// 			// }

// 			// communityPoolExtra := sdk.NewCoins(
// 			// 	sdk.NewCoin(
// 			// 		genesisParams.NativeCoinMetadata.Base,
// 			// 		genesisParams.AirdropSupply,
// 			// 	),
// 			// ).Sub(sumAirdrop)

// 			// fmt.Printf("community pool amount: %s\n", communityPoolExtra)

// 			// distributionGenState.FeePool.CommunityPool = sdk.NewDecCoinsFromCoins(communityPoolExtra...)
// 			// distributionGenStateBz, err := cdc.MarshalJSON(&distributionGenState)
// 			// if err != nil {
// 			// 	return fmt.Errorf("failed to marshal distribution genesis state: %w", err)
// 			// }
// 			// appState[distributiontypes.ModuleName] = distributionGenStateBz

// 			// save entire genesis state to json

// 			appStateJSON, err := json.Marshal(appState)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal application genesis state: %w", err)
// 			}
// 			genDoc.AppState = appStateJSON

// 			err = genutil.ExportGenesisFile(genDoc, genFile)
// 			return err
// 		},
// 	}

// 	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
// 	flags.AddQueryFlagsToCmd(cmd)

// 	return cmd
// }
