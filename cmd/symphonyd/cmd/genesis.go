package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v23/app/params"

	incentivestypes "github.com/osmosis-labs/osmosis/v23/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v23/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v23/x/pool-incentives/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// PrepareGenesisCmd returns prepare-genesis cobra Command.
//

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis",
		Short: "Prepare a genesis file with initial setup",
		Long: `Prepare a genesis file with initial setup.
Examples include:
	- Setting module initial params
	- Setting denom metadata
Example:
	symphonyd prepare-genesis mainnet symphony-1
	- Check input genesis:
		file is at ~/.symphonyd/config/genesis.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			// read genesis file
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// get genesis params
			var genesisParams GenesisParams
			network := args[0]
			if network == "testnet" {
				genesisParams = TestnetGenesisParams()
			} else if network == "mainnet" {
				genesisParams = MainnetGenesisParams()
			} else {
				return fmt.Errorf("please choose 'mainnet' or 'testnet'")
			}

			// get genesis params
			chainID := args[1]

			// run Prepare Genesis
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genesisParams, chainID)
			if err != nil {
				return err
			}

			// validate genesis state
			if err = mbm.ValidateGenesis(cdc, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("error validating genesis file: %s", err.Error())
			}

			// save genesis
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			err = genutil.ExportGenesisFile(genDoc, genFile)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genDoc *tmtypes.GenesisDoc, genesisParams GenesisParams, chainID string) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	depCdc := clientCtx.Codec
	cdc := depCdc

	// chain params genesis
	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisParams.GenesisTime

	genDoc.ConsensusParams = genesisParams.ConsensusParams

	// ---
	// staking module genesis
	stakingGenState := stakingtypes.GetGenesisStateFromAppState(depCdc, appState)
	stakingGenState.Params = genesisParams.StakingParams
	stakingGenStateBz, err := cdc.MarshalJSON(stakingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[stakingtypes.ModuleName] = stakingGenStateBz

	// mint module genesis
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = genesisParams.MintParams
	mintGenStateBz, err := cdc.MarshalJSON(mintGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mint genesis state: %w", err)
	}
	appState[minttypes.ModuleName] = mintGenStateBz

	// distribution module genesis
	distributionGenState := distributiontypes.DefaultGenesisState()
	distributionGenState.Params = genesisParams.DistributionParams
	// TODO Set initial community pool
	// distributionGenState.FeePool.CommunityPool = sdk.NewDecCoins()
	distributionGenStateBz, err := cdc.MarshalJSON(distributionGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal distribution genesis state: %w", err)
	}
	appState[distributiontypes.ModuleName] = distributionGenStateBz

	// gov module genesis
	govGenState := govtypesv1.DefaultGenesisState()
	govGenState.DepositParams = genesisParams.GovParams.DepositParams
	govGenState.TallyParams = genesisParams.GovParams.TallyParams
	govGenState.VotingParams = genesisParams.GovParams.VotingParams
	govGenStateBz, err := cdc.MarshalJSON(govGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	appState[govtypes.ModuleName] = govGenStateBz

	// crisis module genesis
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = genesisParams.CrisisConstantFee
	// TODO Set initial community pool
	// distributionGenState.FeePool.CommunityPool = sdk.NewDecCoins()
	crisisGenStateBz, err := cdc.MarshalJSON(crisisGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal crisis genesis state: %w", err)
	}
	appState[crisistypes.ModuleName] = crisisGenStateBz

	// slashing module genesis
	slashingGenState := slashingtypes.DefaultGenesisState()
	slashingGenState.Params = genesisParams.SlashingParams
	slashingGenStateBz, err := cdc.MarshalJSON(slashingGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal slashing genesis state: %w", err)
	}
	appState[slashingtypes.ModuleName] = slashingGenStateBz

	// incentives module genesis
	incentivesGenState := incentivestypes.GetGenesisStateFromAppState(depCdc, appState)
	incentivesGenState.Params = genesisParams.IncentivesGenesis.Params
	incentivesGenState.LockableDurations = genesisParams.IncentivesGenesis.LockableDurations
	incentivesGenState.Gauges = genesisParams.IncentivesGenesis.Gauges
	incentivesGenStateBz, err := cdc.MarshalJSON(incentivesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal incentives genesis state: %w", err)
	}
	appState[incentivestypes.ModuleName] = incentivesGenStateBz

	// epochs module genesis
	epochsGenState := epochstypes.DefaultGenesis()
	epochsGenState.Epochs = genesisParams.Epochs
	epochsGenStateBz, err := cdc.MarshalJSON(epochsGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal epochs genesis state: %w", err)
	}
	appState[epochstypes.ModuleName] = epochsGenStateBz

	// poolincentives module genesis
	poolincentivesGenState := &genesisParams.PoolIncentivesGenesis
	poolincentivesGenStateBz, err := cdc.MarshalJSON(poolincentivesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal poolincentives genesis state: %w", err)
	}
	appState[poolincentivestypes.ModuleName] = poolincentivesGenStateBz

	// return appState and genDoc
	return appState, genDoc, nil
}

type GenesisParams struct {
	AirdropSupply osmomath.Int

	StrategicReserveAccounts []banktypes.Balance

	ConsensusParams *tmtypes.ConsensusParams

	GenesisTime         time.Time
	NativeCoinMetadatas []banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypesv1.Params

	CrisisConstantFee sdk.Coin

	SlashingParams    slashingtypes.Params
	IncentivesGenesis incentivestypes.GenesisState

	PoolIncentivesGenesis poolincentivestypes.GenesisState

	Epochs []epochstypes.EpochInfo
}

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	genParams.AirdropSupply = osmomath.NewIntWithDecimal(5, 13)           // 5*10^13 note, 5*10^7 (50 million) melody
	genParams.GenesisTime = time.Date(2021, 6, 18, 17, 0, 0, 0, time.UTC) // Jun 18, 2021 - 17:00 UTC

	genParams.NativeCoinMetadatas = []banktypes.Metadata{
		{
			Description: "The native token of Symphony",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    appParams.BaseCoinUnit,
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    appParams.HumanCoinUnit,
					Exponent: appParams.MelodyExponent,
					Aliases:  nil,
				},
			},
			Base:    appParams.BaseCoinUnit,
			Display: appParams.HumanCoinUnit,
		},
		{
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "uion",
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    "ion",
					Exponent: 6,
					Aliases:  nil,
				},
			},
			Base:    "uion",
			Display: "ion",
		},
	}

	genParams.StrategicReserveAccounts = []banktypes.Balance{
		{
			Address: "symphony13gxuc2lq95knh6jdhtpwqnefaalm8jm5uc8rf2",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(47_874_500_000_000))), // 47.8745 million OSMO
		},
		{
			Address: "symphony1wfql9f663yamule5hf5tng0pv9jajqs72xuvc5",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(500_000_000_000))), // 500 thousand OSMO
		},
		{
			Address: "symphony1zn4h42w7qrdj9sc4ks4gaznxs59mh94y7v85md",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(1_000_000_000_000))), // 1 million OSMO
		},
		{
			Address: "symphony1m2y8f4nuy5en8yuu26n0k36awt5f7xhur70xp0",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1mlenhw5yq545rqq5tz6cq4vysth08p46z2j92c",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1d4pdwx0lhq2t4sdhp9dnv7pn2y6cykppdfnq0e",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1gr6r69tkcnuj45ghh2mzfn5d6gw47sgjecppct",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1erwufvjxr7wqe78mhpf5783p842t2cgzag7vuc",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony19437d0nely4tyypttzlz7vg7h4340m92xe3skm",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(25_000_000_000))),
		},
		{
			Address: "symphony1q3d3l0ktm5yggrrcpv8hs86nlsu2jw48a4gghg",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony18tlakavhx3ky4l49a2ejz7uzel6c83addhsrs6",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1z37htkxl5ext0l8uswuc0yrcv7jmhyvqn5mcls",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(25_000_000_000))),
		},
		{
			Address: "symphony1a0yzkcxkqh6q3zg6r8a4eu8acgc064dxzqcvyz",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(5_000_000_000))),
		},
		{
			Address: "symphony1fj3yt7fwgxhj0pux0rp7a8jzzwcku6lndp43kf",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(10_000_000_000))),
		},
		{
			Address: "symphony1ke47rknwfrvzaee2j7p4getywpd8nzdnw94ymf",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(7_500_000_000))),
		},
		{
			Address: "symphony1664jdhg2j5crgyn2n55qq3lj8sq54m8x072hwv",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony12nhzqp4hgpeu4cp8ml4drt6qj2tqavve54qu0g",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1ed9r2p54dnr5f0v2t5r9hwuej0wezavhylgg9u",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1ruhpjhup4s948t20vw62lzvcm8ad7l5dn2l8v6",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(50_000_000_000))),
		},
		{
			Address: "symphony1vg6euqlncdfmk0hu5qmq56c064w3khnp8q8y92",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(1_000_000_000))),
		},
		{
			Address: "symphony13e4v3lu6h7gw4w72r7mg8dr9k8p0xexdgdfucf",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(1_000_000_000))),
		},
		{
			Address: "symphony1mlf5gyk7wzhmrtjlqxzrgtxjk0528hh3qnn0u5",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, osmomath.NewInt(1_000_000_000))),
		},
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.StakingParams.MinCommissionRate = osmomath.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "day"                                                     // 1 day
	genParams.MintParams.GenesisEpochProvisions = osmomath.NewDec(300_000_000_000_000).QuoInt64(365) // 300M * 10^6 / 365 = ~821917.8082191781 * 10^6
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.MintParams.ReductionFactor = osmomath.NewDec(2).QuoInt64(3) // 2/3
	genParams.MintParams.ReductionPeriodInEpochs = 365                    // 1 year (screw leap years)
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          osmomath.MustNewDecFromStr("0.25"), // 25%
		DeveloperRewards: osmomath.MustNewDecFromStr("0.25"), // 25%
		PoolIncentives:   osmomath.MustNewDecFromStr("0.45"), // 45%
		CommunityPool:    osmomath.MustNewDecFromStr("0.05"), // 5%
	}
	genParams.MintParams.MintingRewardsDistributionStartEpoch = 1
	genParams.MintParams.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: "symphony1u7ryvx794sy5yqwezfryygsce84q287ts98n66",
			Weight:  osmomath.MustNewDecFromStr("0.2887"),
		},
		{
			Address: "symphony1zrmuw4xux344w4k9pw93qs8d0d7kc0fnhxw4wd",
			Weight:  osmomath.MustNewDecFromStr("0.2290"),
		},
		{
			Address: "symphony1t9vjrxn6cwdkuf990sncq7akqsz26feaz5euxt",
			Weight:  osmomath.MustNewDecFromStr("0.1625"),
		},
		{
			Address: "symphony172qywhy2qxcnkvr6vcal23ntz645h20qe5880r",
			Weight:  osmomath.MustNewDecFromStr("0.109"),
		},
		{
			Address: "symphony195ds5rrxcqcwflj692e6gmykhl9vu0r0qs7tt5",
			Weight:  osmomath.MustNewDecFromStr("0.0995"),
		},
		{
			Address: "symphony1f2jp2q4qq0f8nlmp0v3ah96h3kqjj0vheprf7q",
			Weight:  osmomath.MustNewDecFromStr("0.06"),
		},
		{
			Address: "symphony1k27t46ehr7y80ktrtmn9grmc9wkw27ds9hq005",
			Weight:  osmomath.MustNewDecFromStr("0.015"),
		},
		{
			Address: "symphony1dhtgp9726rx5zv9079xz2wz43pec484akwktn5",
			Weight:  osmomath.MustNewDecFromStr("0.01"),
		},
		{
			Address: "symphony1fqqucy9y2adaapyjze5g0hv40vp6rt2kt0cjts",
			Weight:  osmomath.MustNewDecFromStr("0.0075"),
		},
		{
			Address: "symphony192953mpz44nn76vgmknt75vspsnv2k6d9dyc4w",
			Weight:  osmomath.MustNewDecFromStr("0.007"),
		},
		{
			Address: "symphony1jcchx5enuex05al39y25gl6hyerwj74unntaqx",
			Weight:  osmomath.MustNewDecFromStr("0.005"),
		},
		{
			Address: "symphony1pt2knp6s8exw7j28gjgmwr2wvw4suc3w8ncunl",
			Weight:  osmomath.MustNewDecFromStr("0.0025"),
		},
		{
			Address: "symphony1c4zx9pmtn3j4a2eus2mmpclpllpqzgzezte7yz",
			Weight:  osmomath.MustNewDecFromStr("0.0025"),
		},
		{
			Address: "symphony1d6fwytjdlwzg7hg26zpzrl4y3f5ykft9xetlmk",
			Weight:  osmomath.MustNewDecFromStr("0.001"),
		},
		{
			Address: "symphony1gmyrqx37tvpmqpkvga6ex4jtv0920hfa3pndqz",
			Weight:  osmomath.MustNewDecFromStr("0.0008"),
		},
	}

	genParams.DistributionParams = distributiontypes.DefaultParams()
	genParams.DistributionParams.CommunityTax = osmomath.MustNewDecFromStr("0")
	genParams.DistributionParams.WithdrawAddrEnabled = true

	genParams.GovParams = govtypesv1.DefaultParams()
	genParams.GovParams.DepositParams.MaxDepositPeriod = time.Hour * 24 * 14 // 2 weeks
	genParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(2_500_000_000),
	))
	genParams.GovParams.TallyParams.Quorum = osmomath.MustNewDecFromStr("0.2") // 20%
	genParams.GovParams.VotingParams.VotingPeriod = time.Hour * 24 * 3         // 3 days

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(500_000_000_000),
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                            // 30000 blocks (~41 hr at 5 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = osmomath.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Minute                           // 1 minute jail period
	genParams.SlashingParams.SlashFractionDoubleSign = osmomath.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = osmomath.ZeroDec()                   // 0% liveness slashing

	genParams.Epochs = epochstypes.DefaultGenesis().Epochs
	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.IncentivesGenesis = *incentivestypes.DefaultGenesis()
	genParams.IncentivesGenesis.Params.DistrEpochIdentifier = "day"
	genParams.IncentivesGenesis.LockableDurations = []time.Duration{
		time.Hour * 24,      // 1 day
		time.Hour * 24 * 7,  // 7 day
		time.Hour * 24 * 14, // 14 days
	}

	genParams.ConsensusParams = tmtypes.DefaultConsensusParams()
	genParams.ConsensusParams.Block.MaxBytes = 5 * 1024 * 1024
	genParams.ConsensusParams.Block.MaxGas = 6_000_000
	genParams.ConsensusParams.Evidence.MaxAgeDuration = genParams.StakingParams.UnbondingTime
	genParams.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(genParams.StakingParams.UnbondingTime.Seconds()) / 3
	genParams.ConsensusParams.Version.App = 1

	genParams.PoolIncentivesGenesis = *poolincentivestypes.DefaultGenesisState()
	genParams.PoolIncentivesGenesis.Params.MintedDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations
	genParams.PoolIncentivesGenesis.DistrInfo = &poolincentivestypes.DistrInfo{
		TotalWeight: osmomath.NewInt(1000),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: 0,
				Weight:  osmomath.NewInt(1000),
			},
		},
	}

	return genParams
}

func TestnetGenesisParams() GenesisParams {
	genParams := MainnetGenesisParams()

	genParams.GenesisTime = time.Now()

	genParams.Epochs = append(genParams.Epochs, epochstypes.EpochInfo{
		Identifier:            "15min",
		StartTime:             time.Time{},
		Duration:              15 * time.Minute,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	})

	for _, epoch := range genParams.Epochs {
		epoch.StartTime = genParams.GenesisTime
	}

	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks

	genParams.MintParams.EpochIdentifier = "15min"     // 15min
	genParams.MintParams.ReductionPeriodInEpochs = 192 // 2 days

	genParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(1000000), // 1 OSMO
	))
	genParams.GovParams.TallyParams.Quorum = osmomath.MustNewDecFromStr("0.0000000001") // 0.00000001%
	genParams.GovParams.VotingParams.VotingPeriod = time.Second * 300                   // 300 seconds

	genParams.IncentivesGenesis = *incentivestypes.DefaultGenesis()
	genParams.IncentivesGenesis.Params.DistrEpochIdentifier = "15min"
	genParams.IncentivesGenesis.LockableDurations = []time.Duration{
		time.Minute * 30, // 30 min
		time.Hour * 1,    // 1 hour
		time.Hour * 2,    // 2 hours
	}

	genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations

	return genParams
}
