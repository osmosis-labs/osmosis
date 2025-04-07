package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	FlagSelectPoolIds      = "breakdown-by-pool-ids"
	FlagMinimumStakeAmount = "minimum-stake-amount"
)

type DeriveSnapshot struct {
	NumberAccounts uint64                    `json:"num_accounts"`
	Accounts       map[string]DerivedAccount `json:"accounts"`
}

// DerivedAccount provide fields of snapshot per account
// It is the simplified struct we are presenting in this 'balances from state export' snapshot for people.
type DerivedAccount struct {
	// TODO: Consider removing this, since duplicated
	Address             string               `json:"address"`
	LiquidBalances      sdk.Coins            `json:"liquid_balance"`
	Staked              osmomath.Int         `json:"staked"`
	UnbondingStake      osmomath.Int         `json:"unbonding_stake"`
	Bonded              sdk.Coins            `json:"bonded"`
	BondedBySelectPools map[uint64]sdk.Coins `json:"bonded_by_select_pools"`
	TotalBalances       sdk.Coins            `json:"total_balances"`
}

// newDerivedAccount returns a new derived account.
func newDerivedAccount(address string) DerivedAccount {
	return DerivedAccount{
		Address:        address,
		LiquidBalances: sdk.Coins{},
		Staked:         osmomath.ZeroInt(),
		UnbondingStake: osmomath.ZeroInt(),
		Bonded:         sdk.Coins{},
	}
}

// underlyingCoins returns liquidity pool's underlying coin balances.
func underlyingCoins(originCoins sdk.Coins, pools map[string]gammtypes.CFMMPoolI) sdk.Coins {
	balances := sdk.Coins{}
	convertAgain := false
	for _, coin := range originCoins {
		if pools[coin.Denom] != nil {
			pool := pools[coin.Denom]
			assets := pool.GetTotalPoolLiquidity(sdk.Context{})
			for _, asset := range assets {
				balances = balances.Add(sdk.NewCoin(asset.Denom, asset.Amount.Mul(coin.Amount).Quo(pool.GetTotalShares())))
				if pools[asset.Denom] != nil { // this happens when there's a pool for LP token swap
					convertAgain = true
				}
			}
		} else {
			balances = balances.Add(coin)
		}
	}

	if convertAgain {
		return underlyingCoins(balances, pools)
	}
	return balances
}

// pools is a map from LP share string -> pool.
// TODO: Make a separate type for this.
func underlyingCoinsForSelectPools(
	originCoins sdk.Coins,
	pools map[string]gammtypes.CFMMPoolI,
	selectPoolIDs []uint64,
) map[uint64]sdk.Coins {
	balancesByPool := make(map[uint64]sdk.Coins)

	for _, coin := range originCoins {
		isLpShare := pools[coin.Denom] != nil
		if !isLpShare {
			continue
		}
		pool := pools[coin.Denom]
		coinPoolID := pool.GetId()

		isSelectPoolID := false
		// check if poolID in select pool IDs
		// TODO: Later change selectPoolIDs to be a hashmap for convenience
		for _, selectID := range selectPoolIDs {
			if selectID == coinPoolID {
				isSelectPoolID = true
				break
			}
		}

		if !isSelectPoolID {
			continue
		}

		// at this point, we've determined this is an LP share for a pool we care about
		balancesByPool[coinPoolID] = underlyingCoins(sdk.Coins{coin}, pools)
	}

	return balancesByPool
}

// getGenStateFromPath returns a JSON genState message from inputted path.
func getGenStateFromPath(genesisFilePath string) (map[string]json.RawMessage, error) {
	genState := make(map[string]json.RawMessage)

	genesisFile, err := os.Open(filepath.Clean(genesisFilePath))
	if err != nil {
		return genState, err
	}
	defer genesisFile.Close()

	byteValue, _ := io.ReadAll(genesisFile)

	var doc tmtypes.GenesisDoc
	err = tmjson.Unmarshal(byteValue, &doc)
	if err != nil {
		return genState, err
	}

	err = json.Unmarshal(doc.AppState, &genState)
	if err != nil {
		panic(err)
	}
	return genState, nil
}

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json.
func ExportDeriveBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-derive-balances [input-genesis-file] [output-snapshot-json]",
		Short: "Export a derive balances from a provided genesis export",
		Long: `Export a derive balances from a provided genesis export
Example:
	symphonyd export-derive-balances ../genesis.json ../snapshot.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			genesisFile := args[0]
			genState, err := getGenStateFromPath(genesisFile)
			if err != nil {
				return err
			}
			snapshotOutput := args[1]

			// Get select bonded pool IDs from flag if its provided
			selectPoolIdsStr, err := cmd.Flags().GetString(FlagSelectPoolIds)
			if err != nil {
				return err
			}
			selectBondedPoolIDs := []uint64{}
			if selectPoolIdsStr != "" {
				selectBondedPoolIDs, err = osmoutils.ParseUint64SliceFromString(selectPoolIdsStr, ",")
				if err != nil {
					return err
				}
			}

			// Produce the map of address to total atom balance, both staked and UnbondingStake
			snapshotAccs := make(map[string]DerivedAccount)

			bankGenesis := banktypes.GenesisState{}
			if len(genState["bank"]) > 0 {
				clientCtx.Codec.MustUnmarshalJSON(genState["bank"], &bankGenesis)
			}
			for _, balance := range bankGenesis.Balances {
				address := balance.Address
				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				acc.LiquidBalances = balance.Coins
				snapshotAccs[address] = acc
			}

			stakingGenesis := stakingtypes.GenesisState{}
			if len(genState["staking"]) > 0 {
				clientCtx.Codec.MustUnmarshalJSON(genState["staking"], &stakingGenesis)
			}
			for _, unbonding := range stakingGenesis.UnbondingDelegations {
				address := unbonding.DelegatorAddress
				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				unbondingMelodys := osmomath.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingMelodys = unbondingMelodys.Add(entry.Balance)
				}

				acc.UnbondingStake = acc.UnbondingStake.Add(unbondingMelodys)

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
					acc = newDerivedAccount(address)
				}

				val := validators[delegation.ValidatorAddress]
				stakedMelodys := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.Staked = acc.Staked.Add(stakedMelodys)

				snapshotAccs[address] = acc
			}

			lockupGenesis := lockuptypes.GenesisState{}
			if len(genState["lockup"]) > 0 {
				clientCtx.Codec.MustUnmarshalJSON(genState["lockup"], &lockupGenesis)
			}
			for _, lock := range lockupGenesis.Locks {
				address := lock.Owner

				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				acc.Bonded = acc.Bonded.Add(lock.Coins...)
				snapshotAccs[address] = acc
			}

			gammGenesis := gammtypes.GenesisState{}
			if len(genState["gamm"]) > 0 {
				clientCtx.Codec.MustUnmarshalJSON(genState["gamm"], &gammGenesis)
			}

			// collect gamm pools
			pools := make(map[string]gammtypes.CFMMPoolI)
			for _, any := range gammGenesis.Pools {
				var pool gammtypes.CFMMPoolI
				err := clientCtx.InterfaceRegistry.UnpackAny(any, &pool)
				if err != nil {
					panic(err)
				}
				pools[gammtypes.GetPoolShareDenom(pool.GetId())] = pool
			}

			// convert balances to underlying coins and sum up balances to total balance
			for addr, account := range snapshotAccs {
				// All pool shares are in liquid balances OR bonded balances (locked),
				// therefore underlyingCoinsForSelectPools on liquidBalances + bondedBalances
				// will include everything that is in one of those two pools.
				account.BondedBySelectPools = underlyingCoinsForSelectPools(
					account.LiquidBalances.Add(account.Bonded...), pools, selectBondedPoolIDs)
				account.LiquidBalances = underlyingCoins(account.LiquidBalances, pools)
				account.Bonded = underlyingCoins(account.Bonded, pools)
				account.TotalBalances = sdk.NewCoins().
					Add(account.LiquidBalances...).
					Add(sdk.NewCoin(appparams.BaseCoinUnit, account.Staked)).
					Add(sdk.NewCoin(appparams.BaseCoinUnit, account.UnbondingStake)).
					Add(account.Bonded...)
				snapshotAccs[addr] = account
			}

			snapshot := DeriveSnapshot{
				NumberAccounts: uint64(len(snapshotAccs)),
				Accounts:       snapshotAccs,
			}

			fmt.Printf("# accounts: %d\n", len(snapshotAccs))

			// export snapshot json
			snapshotJSON, err := json.MarshalIndent(snapshot, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}

			err = os.WriteFile(snapshotOutput, snapshotJSON, 0o644)
			return err
		},
	}

	cmd.Flags().String(FlagSelectPoolIds, "",
		"Output a special breakdown for amount LP'd to the provided pools. Usage --breakdown-by-pool-ids=1,2,605")

	return cmd
}

// StakedToCSVCmd generates a airdrop.csv from a provided exported balances.json.
func StakedToCSVCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "staked-to-csv [input-balances-file] [output-airdrop-csv]",
		Short: "Export a airdrop csv from a provided balances export",
		Long: `Export a airdrop csv from a provided balances export (from export-derive-balances)
Example:
	symphonyd staked-to-csv ../balances.json ../airdrop.csv
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			balancesFile := args[0]

			snapshotOutput := args[1]

			minStakeAmount, _ := cmd.Flags().GetInt64(FlagMinimumStakeAmount)

			var deriveSnapshot DeriveSnapshot

			sourceFile, err := os.Open(balancesFile)
			if err != nil {
				return err
			}
			// remember to close the file at the end of the function
			defer sourceFile.Close()

			// decode the balances json file into the struct array
			if err := json.NewDecoder(sourceFile).Decode(&deriveSnapshot); err != nil {
				return err
			}

			// create a new file to store CSV data
			outputFile, err := os.Create(snapshotOutput)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			// write the header of the CSV file
			writer := csv.NewWriter(outputFile)
			defer writer.Flush()

			header := []string{"address", "staked"}
			if err := writer.Write(header); err != nil {
				return err
			}

			// iterate through all accounts, leave out accounts that do not meet the user provided min stake amount
			for _, r := range deriveSnapshot.Accounts {
				var csvRow []string
				if r.Staked.GT(osmomath.NewInt(minStakeAmount)) {
					csvRow = append(csvRow, r.Address, r.Staked.String())
					if err := writer.Write(csvRow); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64(FlagMinimumStakeAmount, 0, "Specify minimum amount (non inclusive) accounts must stake to be included in airdrop (default: 0)")

	return cmd
}
