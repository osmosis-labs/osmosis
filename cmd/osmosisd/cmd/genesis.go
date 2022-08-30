package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

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
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	appParams "github.com/osmosis-labs/osmosis/v11/app/params"

	epochstypes "osmosis.io/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v11/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v11/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

//nolint:ineffassign
// PrepareGenesisCmd returns prepare-genesis cobra Command.
func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis",
		Short: "Prepare a genesis file with initial setup",
		Long: `Prepare a genesis file with initial setup.
Examples include:
	- Setting module initial params
	- Setting denom metadata
Example:
	osmosisd prepare-genesis mainnet osmosis-1
	- Check input genesis:
		file is at ~/.osmosisd/config/genesis.json
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
	govGenState := govtypes.DefaultGenesisState()
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
	AirdropSupply sdk.Int

	StrategicReserveAccounts []banktypes.Balance

	ConsensusParams *tmproto.ConsensusParams

	GenesisTime         time.Time
	NativeCoinMetadatas []banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypes.Params

	CrisisConstantFee sdk.Coin

	SlashingParams    slashingtypes.Params
	IncentivesGenesis incentivestypes.GenesisState

	PoolIncentivesGenesis poolincentivestypes.GenesisState

	Epochs []epochstypes.EpochInfo
}

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	genParams.AirdropSupply = sdk.NewIntWithDecimal(5, 13)                // 5*10^13 uosmo, 5*10^7 (50 million) osmo
	genParams.GenesisTime = time.Date(2021, 6, 18, 17, 0, 0, 0, time.UTC) // Jun 18, 2021 - 17:00 UTC

	genParams.NativeCoinMetadatas = []banktypes.Metadata{
		{
			Description: "The native token of Osmosis",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    appParams.BaseCoinUnit,
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    appParams.HumanCoinUnit,
					Exponent: appParams.OsmoExponent,
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
			Address: "osmo1el3aytvehpvxw2ymmya4kdyj9yndyy47fw5zh4",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(47_874_500_000_000))), // 47.8745 million OSMO
		},
		{
			Address: "osmo1g7rp8h6wzekjjy8n6my8za3vg3338eqz3v295v",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(500_000_000_000))), // 500 thousand OSMO
		},
		{
			Address: "osmo1z7ql0vcjlznrruw3hwgm043w8yhzqmtpu5rlp9",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(1_000_000_000_000))), // 1 million OSMO
		},
		{
			Address: "osmo1grgelyng2v6v3t8z87wu3sxgt9m5s03xytvfcl",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1yuz4fct5cc3td6fp7he055mut0d4ukcsvlgpkg",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1epr7gp5gf9ftkduzlkckmehqk9dyn063xref53",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1enu76xfqkadu022hkdg7rwyc4pqq6jzn3zy395",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1njvfq5pdad2f5j0uvqtfdujyg8c5sde2n6yw2t",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1gaajper2qphnk6rnpy8zjhhg2kstnd87sh36hx",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(25_000_000_000))),
		},
		{
			Address: "osmo1xl6r0n34mq0z7gfx7w2qf3f2vq70g5ph7n78hs",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo13stzqk27y46cr8fz7vfr8dsqqk4wpz90fqmc23",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo12wmptq2zm3s2d8c0xukwj3fkcl0qw66wmxwr7x",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(25_000_000_000))),
		},
		{
			Address: "osmo1qc74lw7f8ak737jy2at36wa9tyrgldk8kpskzg",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(5_000_000_000))),
		},
		{
			Address: "osmo1m3lvlydlu3mqv7z9xgtqjplcx20yc3lmx3p73e",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(10_000_000_000))),
		},
		{
			Address: "osmo14u94qvqqdyjruufc08m6rwq8yz3ukurkzkvczx",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(7_500_000_000))),
		},
		{
			Address: "osmo15t8y5mzdmz54fcj2wq3qutgtharuxdg7zu9q98",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1ngt8m8722qxcgdpqjwe3ggwcpg6zns8axpxtt0",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo17ncgp84x5vmphdlaaktzp7hljh3mhysqw9up3u",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo16n7070n4whce0wlu76j42dyrxh9f7nlapg6c4a",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(50_000_000_000))),
		},
		{
			Address: "osmo1vnyc6q49sr0hs9ddjepcmtlaq3l6wwj0rrw6hd",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(1_000_000_000))),
		},
		{
			Address: "osmo1ujx3rqerqdksnxnjka2n9tde874mzut75hzx92",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(1_000_000_000))),
		},
		{
			Address: "osmo1a2r7nqnc9e032wj37ptskh207e232462vcrrjf",
			Coins:   sdk.NewCoins(sdk.NewCoin(genParams.NativeCoinMetadatas[0].Base, sdk.NewInt(1_000_000_000))),
		},
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.StakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "day"                                                // 1 day
	genParams.MintParams.GenesisEpochProvisions = sdk.NewDec(300_000_000_000_000).QuoInt64(365) // 300M * 10^6 / 365 = ~821917.8082191781 * 10^6
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.MintParams.ReductionFactor = sdk.NewDec(2).QuoInt64(3) // 2/3
	genParams.MintParams.ReductionPeriodInEpochs = 365               // 1 year (screw leap years)
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.MustNewDecFromStr("0.25"), // 25%
		DeveloperRewards: sdk.MustNewDecFromStr("0.25"), // 25%
		PoolIncentives:   sdk.MustNewDecFromStr("0.45"), // 45%
		CommunityPool:    sdk.MustNewDecFromStr("0.05"), // 5%
	}
	genParams.MintParams.MintingRewardsDistributionStartEpoch = 1
	genParams.MintParams.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
			Weight:  sdk.MustNewDecFromStr("0.2887"),
		},
		{
			Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
			Weight:  sdk.MustNewDecFromStr("0.2290"),
		},
		{
			Address: "osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
			Weight:  sdk.MustNewDecFromStr("0.1625"),
		},
		{
			Address: "osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
			Weight:  sdk.MustNewDecFromStr("0.109"),
		},
		{
			Address: "osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
			Weight:  sdk.MustNewDecFromStr("0.0995"),
		},
		{
			Address: "osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
			Weight:  sdk.MustNewDecFromStr("0.06"),
		},
		{
			Address: "osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
			Weight:  sdk.MustNewDecFromStr("0.015"),
		},
		{
			Address: "osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
			Weight:  sdk.MustNewDecFromStr("0.01"),
		},
		{
			Address: "osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
			Weight:  sdk.MustNewDecFromStr("0.0075"),
		},
		{
			Address: "osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
			Weight:  sdk.MustNewDecFromStr("0.007"),
		},
		{
			Address: "osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
			Weight:  sdk.MustNewDecFromStr("0.005"),
		},
		{
			Address: "osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
			Weight:  sdk.MustNewDecFromStr("0.0025"),
		},
		{
			Address: "osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
			Weight:  sdk.MustNewDecFromStr("0.0025"),
		},
		{
			Address: "osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
			Weight:  sdk.MustNewDecFromStr("0.001"),
		},
		{
			Address: "osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
			Weight:  sdk.MustNewDecFromStr("0.0008"),
		},
	}

	genParams.DistributionParams = distributiontypes.DefaultParams()
	genParams.DistributionParams.BaseProposerReward = sdk.MustNewDecFromStr("0.01")
	genParams.DistributionParams.BonusProposerReward = sdk.MustNewDecFromStr("0.04")
	genParams.DistributionParams.CommunityTax = sdk.MustNewDecFromStr("0")
	genParams.DistributionParams.WithdrawAddrEnabled = true

	genParams.GovParams = govtypes.DefaultParams()
	genParams.GovParams.DepositParams.MaxDepositPeriod = time.Hour * 24 * 14 // 2 weeks
	genParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdk.NewInt(2_500_000_000),
	))
	genParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.2") // 20%
	genParams.GovParams.VotingParams.VotingPeriod = time.Hour * 24 * 3    // 3 days

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdk.NewInt(500_000_000_000),
	)

	genParams.SlashingParams = slashingtypes.DefaultParams()
	genParams.SlashingParams.SignedBlocksWindow = int64(30000)                       // 30000 blocks (~41 hr at 5 second blocks)
	genParams.SlashingParams.MinSignedPerWindow = sdk.MustNewDecFromStr("0.05")      // 5% minimum liveness
	genParams.SlashingParams.DowntimeJailDuration = time.Minute                      // 1 minute jail period
	genParams.SlashingParams.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5% double sign slashing
	genParams.SlashingParams.SlashFractionDowntime = sdk.ZeroDec()                   // 0% liveness slashing

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
	genParams.ConsensusParams.Version.AppVersion = 1

	genParams.PoolIncentivesGenesis = *poolincentivestypes.DefaultGenesisState()
	genParams.PoolIncentivesGenesis.Params.MintedDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations
	genParams.PoolIncentivesGenesis.DistrInfo = &poolincentivestypes.DistrInfo{
		TotalWeight: sdk.NewInt(1000),
		Records: []poolincentivestypes.DistrRecord{
			{
				GaugeId: 0,
				Weight:  sdk.NewInt(1000),
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
		sdk.NewInt(1000000), // 1 OSMO
	))
	genParams.GovParams.TallyParams.Quorum = sdk.MustNewDecFromStr("0.0000000001") // 0.00000001%
	genParams.GovParams.VotingParams.VotingPeriod = time.Second * 300              // 300 seconds

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
