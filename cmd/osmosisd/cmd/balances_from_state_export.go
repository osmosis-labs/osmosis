package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/v8/app/params"
	"github.com/osmosis-labs/osmosis/v8/osmoutils"
	claimtypes "github.com/osmosis-labs/osmosis/v8/x/claim/types"
	gammtypes "github.com/osmosis-labs/osmosis/v8/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

const FlagSelectPoolIds = "breakdown-by-pool-ids"

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
	Staked              sdk.Int              `json:"staked"`
	UnbondingStake      sdk.Int              `json:"unbonding_stake"`
	Bonded              sdk.Coins            `json:"bonded"`
	BondedBySelectPools map[uint64]sdk.Coins `json:"bonded_by_select_pools"`
	UnclaimedAirdrop    sdk.Coins            `json:"unclaimed_airdrop"`
	TotalBalances       sdk.Coins            `json:"total_balances"`
}

func newDerivedAccount(address string) DerivedAccount {
	return DerivedAccount{
		Address:          address,
		LiquidBalances:   sdk.Coins{},
		Staked:           sdk.ZeroInt(),
		UnbondingStake:   sdk.ZeroInt(),
		Bonded:           sdk.Coins{},
		UnclaimedAirdrop: sdk.Coins{},
	}
}

func underlyingCoins(originCoins sdk.Coins, pools map[string]gammtypes.PoolI) sdk.Coins {
	balances := sdk.Coins{}
	convertAgain := false
	for _, coin := range originCoins {
		if pools[coin.Denom] != nil {
			pool := pools[coin.Denom]
			assets := pool.GetAllPoolAssets()
			for _, asset := range assets {
				balances = balances.Add(sdk.NewCoin(asset.Token.Denom, asset.Token.Amount.Mul(coin.Amount).Quo(pool.GetTotalShares().Amount)))
				if pools[asset.Token.Denom] != nil { // this happens when there's a pool for LP token swap
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
// TODO: Make a separate type for this
func underlyingCoinsForSelectPools(
	originCoins sdk.Coins,
	pools map[string]gammtypes.PoolI,
	selectPoolIDs []uint64) map[uint64]sdk.Coins {

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

func getGenStateFromPath(genesisFilePath string) (map[string]json.RawMessage, error) {
	genState := make(map[string]json.RawMessage)

	genesisFile, err := os.Open(genesisFilePath)
	if err != nil {
		return genState, err
	}
	defer genesisFile.Close()

	byteValue, _ := ioutil.ReadAll(genesisFile)

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

			authGenesis := authtypes.GenesisState{}
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["auth"], &authGenesis)
			accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
			if err != nil {
				panic(err)
			}
			accounts = authtypes.SanitizeGenesisAccounts(accounts)

			// Produce the map of address to total atom balance, both staked and UnbondingStake
			snapshotAccs := make(map[string]DerivedAccount)

			bankGenesis := banktypes.GenesisState{}
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["bank"], &bankGenesis)
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
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["staking"], &stakingGenesis)
			for _, unbonding := range stakingGenesis.UnbondingDelegations {
				address := unbonding.DelegatorAddress
				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				unbondingOsmos := sdk.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingOsmos = unbondingOsmos.Add(entry.Balance)
				}

				acc.UnbondingStake = acc.UnbondingStake.Add(unbondingOsmos)

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
				stakedOsmos := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.Staked = acc.Staked.Add(stakedOsmos)

				snapshotAccs[address] = acc
			}

			lockupGenesis := lockuptypes.GenesisState{}
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["lockup"], &lockupGenesis)
			for _, lock := range lockupGenesis.Locks {
				address := lock.Owner

				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				acc.Bonded = acc.Bonded.Add(lock.Coins...)
				snapshotAccs[address] = acc
			}

			claimGenesis := claimtypes.GenesisState{}
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["claim"], &claimGenesis)
			for _, record := range claimGenesis.ClaimRecords {
				address := record.Address

				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				claimablePerAction := sdk.Coins{}
				for _, coin := range record.InitialClaimableAmount {
					claimablePerAction = claimablePerAction.Add(
						sdk.NewCoin(coin.Denom,
							coin.Amount.QuoRaw(int64(len(claimtypes.Action_name))),
						),
					)
				}

				for action := range claimtypes.Action_name {
					if record.ActionCompleted[action] == false {
						acc.UnclaimedAirdrop = acc.UnclaimedAirdrop.Add(claimablePerAction...)
					}
				}

				snapshotAccs[address] = acc
			}

			gammGenesis := gammtypes.GenesisState{}
			clientCtx.JSONCodec.MustUnmarshalJSON(genState["gamm"], &gammGenesis)

			// collect gamm pools
			pools := make(map[string]gammtypes.PoolI)
			for _, any := range gammGenesis.Pools {
				var pool gammtypes.PoolI
				err := clientCtx.InterfaceRegistry.UnpackAny(any, &pool)
				if err != nil {
					panic(err)
				}
				pools[pool.GetTotalShares().Denom] = pool
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

			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	cmd.Flags().String(FlagSelectPoolIds, "",
		"Output a special breakdown for amount LP'd to the provided pools. Usage --breakdown-by-pool-ids=1,2,605")

	return cmd
}
