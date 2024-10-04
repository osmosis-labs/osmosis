package cmd

import (
	"encoding/json"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	oracletypes "github.com/osmosis-labs/osmosis/v26/x/oracle/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v26/x/txfees/types"
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
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v26/app/params"

	epochstypes "github.com/osmosis-labs/osmosis/v26/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v26/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v26/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v26/x/pool-incentives/types"
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

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genDoc *genutiltypes.AppGenesis, genesisParams GenesisParams, chainID string) (map[string]json.RawMessage, *genutiltypes.AppGenesis, error) {
	depCdc := clientCtx.Codec
	cdc := depCdc

	// chain params genesis
	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisParams.GenesisTime

	genDoc.Consensus.Params = genesisParams.ConsensusParams

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
	govGenState.Params = &genesisParams.GovParams
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

	// txtypes module genesis
	txfeesGenState := &genesisParams.TxFees
	txfeesGenStateBz, err := cdc.MarshalJSON(txfeesGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal txfees genesis state: %w", err)
	}
	appState[txfeestypes.ModuleName] = txfeesGenStateBz

	// auth module
	authGenState := &genesisParams.AuthState
	authGenStateBz, err := cdc.MarshalJSON(authGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}
	appState[authtypes.ModuleName] = authGenStateBz

	// bank module
	bankGenState := &genesisParams.BankState
	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	// oracle module
	oracleGenState := &genesisParams.OracleState
	oracleGenStateBz, err := cdc.MarshalJSON(oracleGenState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal oracle genesis state: %w", err)
	}
	appState[oracletypes.ModuleName] = oracleGenStateBz

	// return appState and genDoc
	return appState, genDoc, nil
}

type GenesisParams struct {
	//AirdropSupply osmomath.Int

	//StrategicReserveAccounts []banktypes.Balance

	ConsensusParams *tmtypes.ConsensusParams

	GenesisTime         time.Time
	NativeCoinMetadatas []banktypes.Metadata

	AuthState          authtypes.GenesisState
	BankState          banktypes.GenesisState
	OracleState        oracletypes.GenesisState
	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypesv1.Params

	CrisisConstantFee sdk.Coin

	SlashingParams    slashingtypes.Params
	IncentivesGenesis incentivestypes.GenesisState

	PoolIncentivesGenesis poolincentivestypes.GenesisState

	Epochs []epochstypes.EpochInfo

	TxFees txfeestypes.GenesisState
}

func MainnetGenesisParams() GenesisParams {
	genParams := GenesisParams{}

	//genParams.AirdropSupply = osmomath.NewIntWithDecimal(5, 13)           // 5*10^13 note, 5*10^7 (50 million) melody
	genParams.GenesisTime = time.Now()

	genParams.BankState = *banktypes.DefaultGenesisState()
	genParams.AuthState = *authtypes.DefaultGenesisState()
	genParams.OracleState = *oracletypes.DefaultGenesisState()

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
			Description: "The native token of Symphony",
		},
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genParams.StakingParams.MaxValidators = 100
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.StakingParams.MinCommissionRate = osmomath.MustNewDecFromStr("0.05")

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.EpochIdentifier = "week"
	genParams.MintParams.GenesisEpochProvisions = osmomath.NewDec(150250000000)
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.MintParams.ReductionFactor = osmomath.NewDecWithPrec(5, 1) // 0.5
	genParams.MintParams.ReductionPeriodInEpochs = 208                   // 4 years
	genParams.MintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          osmomath.NewDecWithPrec(7, 1),  // 0.7
		PoolIncentives:   osmomath.NewDec(0),             // 0.0
		DeveloperRewards: osmomath.NewDecWithPrec(25, 2), // 0.25
		CommunityPool:    osmomath.NewDecWithPrec(5, 2),  // 0.05
	}
	genParams.MintParams.MintingRewardsDistributionStartEpoch = 1
	genParams.MintParams.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: "symphony1g6pxgl8g0rnk7q86j9zh7yxsqdsn7jvdmc8fkx",
			Weight:  osmomath.MustNewDecFromStr("1.0"),
		},
	}

	genParams.DistributionParams = distributiontypes.DefaultParams()
	genParams.DistributionParams.CommunityTax = osmomath.MustNewDecFromStr("0")
	genParams.DistributionParams.WithdrawAddrEnabled = true

	genParams.GovParams = govtypesv1.DefaultParams()
	*genParams.GovParams.MaxDepositPeriod = time.Hour * 24 * 14 // 2 weeks
	genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(2_500_000_000),
	))
	genParams.GovParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		osmomath.NewInt(5_000_000_000),
	))
	genParams.GovParams.Quorum = "0.2"                     // 20%
	*genParams.GovParams.VotingPeriod = time.Hour * 24 * 3 // 3 days

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

	genParams.TxFees = *txfeestypes.DefaultGenesis()
	genParams.TxFees.Basedenom = genParams.NativeCoinMetadatas[0].Base

	// oracle
	defaultTobixTax := osmomath.ZeroDec()
	genParams.OracleState.TobinTaxes = []oracletypes.TobinTax{
		{Denom: appParams.MicroUSDDenom, TobinTax: defaultTobixTax},
		{Denom: appParams.MicroHKDDenom, TobinTax: defaultTobixTax},
		{Denom: appParams.MicroVNDDenom, TobinTax: defaultTobixTax},
	}
	genParams.OracleState.ExchangeRates = oracletypes.ExchangeRateTuples{
		{Denom: appParams.MicroUSDDenom, ExchangeRate: osmomath.NewDecWithPrec(10, 0)},    // 1 USD = 10 MLD
		{Denom: appParams.MicroHKDDenom, ExchangeRate: osmomath.NewDecWithPrec(12820, 4)}, // 1 HKD = 1,2820 MLD
		{Denom: appParams.MicroVNDDenom, ExchangeRate: osmomath.NewDecWithPrec(399, 6)},   // 1 VND = 0,000399 MLD
	}

	return genParams
}

func TestnetGenesisParams() GenesisParams {
	genParams := MainnetGenesisParams()

	genParams.GenesisTime = time.Now()

	//genParams.Epochs = append(genParams.Epochs, epochstypes.EpochInfo{
	//	Identifier:            "15min",
	//	StartTime:             time.Time{},
	//	Duration:              15 * time.Minute,
	//	CurrentEpoch:          0,
	//	CurrentEpochStartTime: time.Time{},
	//	EpochCountingStarted:  false,
	//})
	//
	//for _, epoch := range genParams.Epochs {
	//	epoch.StartTime = genParams.GenesisTime
	//}

	genParams.StakingParams.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks

	//genParams.MintParams.EpochIdentifier = "15min"     // 15min
	//genParams.MintParams.ReductionPeriodInEpochs = 192 // 2 days
	//
	//genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
	//	genParams.NativeCoinMetadatas[0].Base,
	//	osmomath.NewInt(1000000), // 1 OSMO
	//))
	//genParams.GovParams.Quorum = "0.0000000001"           // 0.00000001%
	//*genParams.GovParams.VotingPeriod = time.Second * 300 // 300 seconds
	//
	//genParams.IncentivesGenesis = *incentivestypes.DefaultGenesis()
	//genParams.IncentivesGenesis.Params.DistrEpochIdentifier = "15min"
	//genParams.IncentivesGenesis.LockableDurations = []time.Duration{
	//	time.Minute * 30, // 30 min
	//	time.Hour * 1,    // 1 hour
	//	time.Hour * 2,    // 2 hours
	//}
	//
	//genParams.PoolIncentivesGenesis.LockableDurations = genParams.IncentivesGenesis.LockableDurations

	err := createTestnetAccounts(&genParams)
	if err != nil {
		panic(err)
	}

	return genParams
}

func createTestnetAccounts(genParams *GenesisParams) error {
	var accs authtypes.GenesisAccounts
	var bankBalances []banktypes.Balance
	genesisBalances := map[string]sdk.Coins{
		"symphony13luum7djwdhkqg3tw9rae04an6rl036095d7qr": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(10_000_000*appParams.MicroUnit))), // 10 million note
		"symphony1qdvzqujxqd0pqwcdtpxgfcqcvxn777kax7mu86": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(1_000_000*appParams.MicroUnit))),  // 1 million note to faucet
		"symphony1nkdh6l5wkygv7cuw6kfalknpus6fqmsr746f6k": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(1_000*appParams.MicroUnit))),      // 1,000 note to validator
		"symphony16agv74asz2zlcyq4eusk4h0hkxwpp0hxex83jk": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(1_000*appParams.MicroUnit))),      // 1,000 note to validator
		"symphony1dmlepawltn5hrvmz6humx99rrh48jdst4dce86": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(1_000*appParams.MicroUnit))),      // 1,000 note to validator
		"symphony1vhhgnhmnw0x9zfslv4m2plchx8ecgthugemp74": sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(1_000*appParams.MicroUnit))),      // 1,000 note to validator
	}

	createTestnetAirdropAccounts(genesisBalances)

	for addr, coins := range genesisBalances {
		addr := sdk.MustAccAddressFromBech32(addr)
		balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
		baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
		if err := baseAccount.Validate(); err != nil {
			return fmt.Errorf("failed to validate new genesis account: %w", err)
		}
		accs = append(accs, baseAccount)
		bankBalances = append(bankBalances, balances)
	}
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}
	genParams.AuthState.Accounts = genAccs
	genParams.BankState.Balances = banktypes.SanitizeGenesisBalances(bankBalances)

	return nil
}

func createTestnetAirdropAccounts(genesisBalances map[string]sdk.Coins) {
	airDropAccounts := []string{
		"symphony1nhfhxk692c9svf0th9ktlpsfsr6askcr8fs2xv",
		"symphony1ugf25wya496ws8638w6750wjj2qmdh4c3egp38",
		"symphony1044f5uwzmjn87m6uutj8tmvanak7nnz35r8gxa",
		"symphony1quyz2y0449dfg4hzjylr5y4rmf79rsa74hf94m",
		"symphony1gx63mdytxuej9s3lk3n9hkw6mhuhvphr9qqxzj",
		"symphony18ljwl6mctxqlzkf68s3jj37x79ukh64v42mzcl",
		"symphony1v8w8gmc4zahffmkdq3r8ha7nrptuhpcm6jcqzn",
		"symphony1cka5gu53jasmr5dv8j7dwueydna4pfkp8875sl",
		"symphony15fz6rfdkwzy8pwglgmt44ehyr07u38hhv7t023",
		"symphony1gh7sn0ysa5d2rxv5ajwhqatcasljazn2whlzm6",
		"symphony1tf5w397vwmu7424uuvws5rmtaeshavfqcn2eff",
		"symphony1jw770sh9srxzf9gvuvqpu7kdlh2w3nze5xzcga",
		"symphony12c3yc0wpvx5afa3fdgs4eulum4pn8rdna6h945",
		"symphony1fd7m669vm9admcs2scxsgk3jwq3rv6pjjxl7q5",
		"symphony1g3g4270wxgje6npqqr9vxmyxhgnearzaytv0vf",
		"symphony1w7mp6ft6ns9hupmam4nd78940agf49ueqa9nan",
		"symphony1rl8988ugjllepjlcrulajplrpn90w2qshl6ehp",
		"symphony1wpayju4jcn2mhv6yewclf6rcq9fyqzvakdg86h",
		"symphony13ts3j8q27kcykenfzty7lre6gvgh8w627m5tvf",
		"symphony1cf5a4t6ecwrzl056af228v9grsdw2rdrw2e2yx",
		"symphony1ay8r5twk6rju27kxfjcrwxyjgmh79ta7mve3e2",
		"symphony1cmtggvey9qrucfsdx6d3qf8gejczeqkqsxkh62",
		"symphony1vn20hl7y9j0rjruezwewzxlfmh6mxzr5v5npc4",
		"symphony1dlf5c6q26xz5djldfn5j5xqya27d37ar5us4ar",
		"symphony1aydacu8zqhfh6j7tzfz2x6vltdjeql5qerejv8",
		"symphony1q8slpmqh5zaxswgcepw5ae3220hpnasuu33q2w",
		"symphony1uf726gg73h3cvtn438cnugn4gq3vn8arzexcl2",
		"symphony1sfq6mjwvm2vlzg50s998ts2m4xm0wc0wmuvul2",
		"symphony14tv5jf4uzkm8qt5y7rjc58szu33hs6f0kts7la",
		"symphony10nvahvp446zuhrdj9y9vnkduenmjmxmkek4rxz",
		"symphony1fn0vts2t5wlnftcs655ls853rymfnek0yjq4lt",
		"symphony1ykklg2g84gsx2hznx3x4nepepye2h506tcpfna",
		"symphony10khujjn6jdvlg6nqu3kl3jj0jqgkpxrxck39wr",
		"symphony1cxngwchyr078sma8v64npp7783yt03ywgatmx7",
		"symphony1qgz4n5a3kerh8cn5rrcwd8wkleaujhvssygtw8",
		"symphony1088l6ajrxag05d6drcd9arte38xzukzcydq9nm",
		"symphony1cjdxm4urpdp42un8xjsdx6469h3nlx26rewtrl",
		"symphony1ypwdszxftup0punytywgt7fptwsr3zzk0lk892",
		"symphony1vr5trm8e5atgy3r23qufx98zmlk8k36trnwuul",
		"symphony1zrrp94t7pxuc772rm0c4m4rq9pmd25mnuegdvu",
		"symphony1f5kngdadf673gwz7mlahjypq7um9du4jn9q622",
		"symphony1rt483yg05t6ljpmlc55cpmwr28m60ssj654hfk",
		"symphony1svh6t7q3u45se246n7jpcxzgrx0kmyga06a4ee",
		"symphony1rl53j93fl3n2jky0e32q067du3lddagh08y3xf",
		"symphony1mwrvywk725aywlgc8ky68gcpxyw7ww0mrg6gze",
		"symphony128qmxx7hrf5z6yd87m4nkdsn4jrmda62ehtcj0",
		"symphony1swayw9a9wa97gpw4xye37dvektzsfq6hkwpznw",
		"symphony1t75pl6rfd3n2626ww0grhv4qgzvegtde6stg3f",
		"symphony1sdv3fyn7lvrhyuvmtlvn3sr4qp60acp5yaw3h0",
		"symphony1p6glwfxxdmq8x9skh0vw4t4n6q7857d3fc3530",
		"symphony17hw2lcjut7d5h93krmdqw7rtex602v3tpp7qsk",
		"symphony1hel9md4mwxtwfrjjna60fwynfuqe55avw4mask",
		"symphony102h03d76rnhtk9guqkar48vnyplsuq0lqkxqx7",
		"symphony1d273uzgffa8k5406xdj4zllt0wz8u8zkkmdxl8",
		"symphony1q3yw9srwkwelqrue7pzy9anzuy3p386wue0km2",
		"symphony19zc8q349wuzch8r57jmhj9h8xmhrs0rq0zkvyn",
		"symphony1s9krssrnyw3fwtn7uq37m3lhm57gy2gqjgqzxz",
		"symphony18uvf6p8glz04lrgy2k8nudmp90vs26hl3jcffn",
		"symphony1xvyxsc26kv2ct6x5khs3yggah09lleg6q4zrux",
		"symphony1rhnu5rwraxk64tukvrav8vgx8l3sgmhgfxe9u3",
		"symphony1dsmnavqp0nfwjpjcyjypdy4fa9z8pv4v33rera",
		"symphony1qe7knkw3xfrrashn8m7jzphslcpszgyvhdq00d",
		"symphony1ygpgpmf6f82zppetseup8mnt8kjuxx70ratgcj",
		"symphony1tegl6tlcfymyke45lrhvatshz5t4xnn3ptqpej",
		"symphony1rxkkw6gzcym2vxtkqwvagh3ayf7fwhgdyjzdp8",
		"symphony1zswvnruty780ymad4wc6cccgzqf7l524r5wh2l",
		"symphony1pt8hmxla268hq8a48wshhu2ehw0r6xmzl4gcyj",
		"symphony1pkxapj9rzpde8899rr9fzdus3ryyd7nhajzu8q",
		"symphony1rtyehaur9xyjrk4kr8yvvyp7k9vm8jl5rfaqu3",
		"symphony14z0vg5c4nks2lt0kvfmpzdptzj249dqs8herhe",
		"symphony18xjrhmn3pay83rgl04a8guvj5al8zs5fm39wty",
		"symphony167jm5g3sknlaqv2043p326zmx6vwh887qaehe6",
		"symphony1p6l4v0umr4my40ewwjpzxyscfzqg3zyhhs2vnu",
		"symphony1jkywj45pfqqqfjuxkd5r0azrw0ktvdp284rgdp",
		"symphony1wlx03ha5tl50zy4jejxx6q5gx29j7zqfhldfx0",
		"symphony1tpe8a8gc0408z6lz2uq3z74chz5wm7weu29xsh",
		"symphony108j4xd0lkmu3wu9zad9h5gepxq35jpunh99426",
		"symphony1kea6x60nq7pjt62y88x6p8uavvngdn7sya8hf4",
		"symphony1qrgavreluaum2ykpqc3czmxguv0vqey48zjgl7",
		"symphony1ndmygfgadgl965fl4yflzc7s0eq8drgdpg5j4q",
		"symphony1k9kawm9ng0krsqtrhauscwnchwklev33egjrpp",
		"symphony1dttsw5s5nj0g3j0ex0j0l4xn8hys6c5as9x7hz",
		"symphony1mp6tjgjds79np9qssfgn9vaknyjedg05jf0lc8",
		"symphony1n9eg3xays4s7yf0nxq5ke8l6a0awu5t7u0482v",
		"symphony1hn05ra43tes46a99efdf65762v2wfzkuw3h3d5",
		"symphony1qr4chpevu9rd6r2uscz473c2292u7kz02sx9h5",
		"symphony1qtcxr57jsgh3xa9hqsmntc9tmuc8tjejunk3j8",
		"symphony18drfeplqvgtszvjd09sn6hafutc44qm2vnj7x6",
		"symphony1rgud6cg2rrps00fxkant0e6fzgrnpsatt094uf",
		"symphony15vjv50z6vda6djaskjsy3patxg5ha2z0v4paxd",
		"symphony1ju75yt667nef02pts5lvl42gssep54q0whu4aw",
		"symphony1znzcuu2k795pr3ahhn2nptnvt6trfwecd5rdlk",
		"symphony1rmwf8yvesdcs5awxdxy850wa6nufp2344wkkxt",
		"symphony1uuxk5exq5d8gqnudvgk5rkcqe9g0f5qzauln9d",
		"symphony19cx0qpksnn280aqktqqppdvqxsplu5sr364xqz",
		"symphony17m606t2c6ge94afee5m5p7h3h7s5lc7zx2zu8v",
		"symphony1puvf3j9m6a2dr83wa8p8p3ue08we2vr6dcqgs4",
		"symphony17x84cp6laf34uypvxq4nk9f575qsxfrhlag542",
		"symphony1khel0ty6v2wz0tke9nr0k03ce5jdrgvtdjf2rl",
		"symphony1lhg4jj4qmmkkvcujly0lx3x5jnk465v63naan3",
		"symphony18aztuq8rhrw7fjsdnw5zhrfpvcc5vf8qfs840u",
		"symphony1q2lwuepfyrt4mwl8luyqpgu02g8q5k7mpknzqf",
		"symphony1dthzgqmtyhdl7rwj47vuege7zumtx8se9n8z0l",
		"symphony1qap8p5jjhagrqh0q0e9wk5k5atf8z30hmlgqh8",
		"symphony19ua4hvunr4g5rdyzg64x22qstz0mp4kmh0galv",
		"symphony1arna46ll33uecwfzd3elvh4ppwrl7cf6srpcmn",
		"symphony17v07wc7s5ssw3yzp0rnycuygnjzcglvu0mqa60",
		"symphony1nx44yj7dazwl7wqanc67fy0emzsra0wrs3esxf",
		"symphony1czw438l4pu2htyu9qegprfk4057xqvly7tyq8e",
		"symphony1guscruewuwul076qfdwwqddkc50lkersf3dpcr",
		"symphony16d6exmu8snjv2af0z672je5qawr5mcwctc2ul3",
		"symphony1tkun8ehnswrlhnhe4sz6xxdvtnzmxpmdklpjmj",
		"symphony12d4hl9qmdyyg9jequsdts4k7lw03ug86uxxcqp",
		"symphony14ng3qq7cmkvjaqjxdnpn7yx7v6fhmnsty6m8el",
		"symphony13m2wgclv440fdtfa5vrxhfuxtnrcpvcza52zus",
		"symphony12d4hl9qmdyyg9jequsdts4k7lw03ug86uxxcqp",
		"symphony14r6z0duck4dgewev2df0q3rkhmhw8p8cxrnn2d",
		"symphony1u2lqexjczvj5lxspa7rh4ttv8nm6xqqtv2u4q0",
		"symphony1c9l75tg0xaggvpraye0g5t48444h57sskfvtcc",
		"symphony1w3zjrycvqjv7zgywjx5dhuptxllny0mqsumqp2",
		"symphony1w3zjrycvqjv7zgywjx5dhuptxllny0mqsumqp2",
		"symphony1e32l28uzkz2vk6tsauxfmlmw83phu9l5jtta9e",
		"symphony1jsgr0aw2peuxdf0n86r9c8qeawv04q4zgyhmlv",
		"symphony1mn99raxzlkm0trgjedfq28rrarw94n89rk823g",
		"symphony1qfxzqrlglcg5l84ncstdtfx5way306agkqetl2",
		"symphony1qfxzqrlglcg5l84ncstdtfx5way306agkqetl2",
		"symphony144cq3tuq90l5ene0f5uamdd723a3hkdvsp2cp5",
		"symphony122jxy9whr0d3ygdur5kknnce9cwwn0mc0gpmwu",
		"symphony1gz6yelsar8st6nd3l6k3svs782hxzhrvu3jwls",
		"symphony1gvn8hprd6eypvahjrpfxh8mvz84pjw672nlvcf",
		"symphony1gvn8hprd6eypvahjrpfxh8mvz84pjw672nlvcf",
		"symphony1stlklws64a7lmrljk7y30v33wzzpjw3mpp787u",
		"symphony1q057se34vmll90gxjyvx8k6rld5udxm97awm6k",
		"symphony1hferwqruxsc2nn2934r4fhdwg7c6n0auykf4d9",
		"symphony1u2yxjhtgnpvpl959c6ua0vw0hkupakxz5k4qul",
		"symphony19j0ps84zdnd5nqj037cg8f4t3c9fl34j6fv2jq",
		"symphony1pcgtjnwzg3xllp7sr8mcepcfwhtsnv2katrgcw",
		"symphony1selag4pz5d3ak6lve9y9cqxycademd59jpnzjt",
		"symphony1ftj9edf8uwjj6vmckthtl3vdkj82796w29d3uc",
		"symphony1a7el09t9mcu88xy7243dj9pd0nc4f7nyrpn6cx",
		"symphony1tw2eyvpxwlax94nxgpfy5zuuv3fclwyc7jdfwz",
		"symphony18drlcev3ywavxzr4q37at5hf55ugfcdf3yuwxn",
		"symphony1kr09pvscyg5ze98wut2qmkt9ru999exqjc6lef",
		"symphony1lmkfwskcg8qmm42k4l5um5pw49599d2fhfajgh",
		"symphony1j2nfz3wxlq6h0fd2qnngrtnmdwh4uqf73dmqq4",
		"symphony1khel0ty6v2wz0tke9nr0k03ce5jdrgvtdjf2rl",
		"symphony1g62ujdnm6c4rpg2efr34m76usd2s9kr2tg0jj9",
		"symphony10vc0w9tzmggettnm6az040fl6naapc59rlzlxs",
		"symphony1sjfhecqys9vmn3m7q0mckqlnmaq9kwzmuumwdr",
		"symphony17qez6kd3l023jatjsftdj0g9p8604dxq8gk5u9",
		"symphony1xnjytumgfu2fy98uuzm7tny7ljd67r8ddpyw8u",
		"symphony14audndq4wv4xjqc8wwc5mf5x8l50zc06kpnpxa",
		"symphony1wl59rm0fvr2j98tch4dk50gk5529gpr2pdtj76",
		"symphony1vt34g8z57kgaefzfpqxqpc0x82cdpnj7k90v83",
		"symphony1ypherq5ykaj8e7n6dx3gq7tlqz788k3f27jmnp",
		"symphony1cfkeut9pehjk45yc3admczt2dgsrzmc0jzq6ug",
		"symphony1mrptlzxpkpjvwh03auhqh7v92r8x0unx5jnzmw",
		"symphony1nsv3h9dhtxp0alazmvjcddye203lfe6jtzdu6e",
		"symphony160vajtgmepv9jpa65ufxwmtzahfqgzv4maazl8",
		"symphony1vq2lumwfn7muwrcdajepvxh0xxc5a90wl24g5f",
		"symphony1em86lcu26lsj8jek7wvnxw4ffgjcer29hmj7wx",
		"symphony18kqfc54kvr87ak97f83ae7dqe5v9k8pz5250yc",
		"symphony1w6344vg2s7s5frfq9c9h4je80zram944087zt9",
		"symphony1ynwwfz9rm7h5n5s67nlhfppj4pyrcdmqu9klfw",
		"symphony1g2v22vmaxv58vj5fh6zdhdcr3v02cag08t8pfr",
		"symphony1jhxx388tyz58mh8tqrfcgdpnaaj88y9l6dtmmc",
		"symphony174wsluse88shxl2e9kslzf57dmrwq9ztjefqfz",
		"symphony17cnpkqrka4tcnqh4l6dlkkfev4gl5p4c6fs03r",
		"symphony1h7gqha93klq65jfm8p96gfl9y25u8w8hwl0rxw",
		"symphony1q2lf266fjapadqazxqfnjw4vqnymy3v0s5mulr",
		"symphony16670d9yw2fv39rc466dd69zwrdxk5lqs0fta92",
		"symphony1g7u2rvuff7lx73eahtr2qsvngv6ptp2hqmp8jp",
		"symphony1gc5z0xekxegknu2rwuj9m0x4dh4dgmrqyj84fz",
		"symphony17tmlpkdxjed9mt58jwc6ynz6gpf2aqe57kelyy",
		"symphony15gfkc3f46xw4eqe895wvlnqhhmwenzv0wlhz36",
		"symphony1vrfgd73hs7ntm24fmy3kxrjmhr5jjqqcx5g2t9",
		"symphony132e95pg44t6r8du4hf7hsycc4qctete8mh6xqv",
		"symphony1f8kvxgzlysspzsrsq543jexzlvhufeys7slg6s",
		"symphony1wheg4th5e8ru74k8rcypq2zuj4k38n4mhzcfsv",
		"symphony1twaerwccc3z238hfxu35hk9wafx4qkwr43eu3h",
		"symphony16jlgzevhnwd72zcfqwhd3kgw0tmc7e7pth4x70",
		"symphony165guumu29v77prtv9hqjx5yj5t035azzltth0y",
		"symphony1a88l3m9x09utlrhcmp99uraqa4y3xqs5zmk65d",
		"symphony1fesr3ltm89rqnmhygrdd5x3f69cytjdh9nyrpc",
		"symphony1k53vn2un4medxl3etlr8p09pf3532myevkg08y",
		"symphony1twaerwccc3z238hfxu35hk9wafx4qkwr43eu3h",
		"symphony17tmlpkdxjed9mt58jwc6ynz6gpf2aqe57kelyy",
		"symphony16jlgzevhnwd72zcfqwhd3kgw0tmc7e7pth4x70",
		"symphony1fnjttapww3kw7pjdvpjl9v0wf3n5eem7r9qtzy",
		"symphony1qmc2lvrpurr3y0z8s904ufdesa4hnucy2dxj93",
		"symphony1rdc9tv9xzhv3dv23zqdheh4dr7q50pqy94h3vs",
		"symphony1ehasewyq4rszdcev9jhvtnyzkhagy0lad6q2zt",
		"symphony16rqdz9r4ekk0e5u8m9cxqc2r9h8g7nkued2tkn",
		"symphony185lw38n7hm8x8d58dlena4pks6250ja9xdwz23",
		"symphony1fefmaqrd3f9vczv8q5hw8e6enqvmq4q4amtf6w",
		"symphony1unqyqqlcc6p36pwvn5p7qelxp32vrn55zrzeyp",
		"symphony14t2svsux8jafuyem9k898k7d077mrtm0puvc50",
		"symphony1r450878wcrs63rg8dw4wa3d0a9r8qk645xr9yy",
		"symphony1drah4h0ap3ph9eusapfr2077k0qt28ptlecjwl",
		"symphony1y4na37uhjfgdgdp4lgcq5svyfrcldty8q9t9gf",
		"symphony1t75pl6rfd3n2626ww0grhv4qgzvegtde6stg3f",
		"symphony1h7gqha93klq65jfm8p96gfl9y25u8w8hwl0rxw",
		"symphony1vmcfjdj2pwuz7k3qf5unk5c9r347clhp69tz47",
		"symphony1p6glwfxxdmq8x9skh0vw4t4n6q7857d3fc3530",
		"symphony1a88l3m9x09utlrhcmp99uraqa4y3xqs5zmk65d",
		"symphony1fesr3ltm89rqnmhygrdd5x3f69cytjdh9nyrpc",
		"symphony1aytslkp026tn574g3686lufeguycpguukaze22",
		"symphony1ems938rufxyex5yv8zny7j3vc3kns8mrl07pyh",
		"symphony1sj3xcyt05dtynd0fdt36zufndpv9ltmu6wnr25",
		"symphony19k9qjyj00aadm45llngf95z5v73hggz4wmefrr",
		"symphony1u2yxjhtgnpvpl959c6ua0vw0hkupakxz5k4qul",
		"symphony1y393sfzqkva28st6a7tffsaa2er9pg532w6fd7",
		"symphony1lyd4ktmvfjvqcj5x5glep0w7w9y8r48sa066n0",
		"symphony19x2vh75wlv24nkwalp74lu7xjujrfcmyrs2qdx",
		"symphony1qnvcxf2ytu942gx3ejzuzcqmjcs3p3eyeuq2tv",
		"symphony1rh79mzpxf0y3ehjz3ucajv39a75kvpn3sgf0dt",
		"symphony1q2lf266fjapadqazxqfnjw4vqnymy3v0s5mulr",
		"symphony1nlvjat58pl5dgdn5ycf6n38ksy0kzgh3yp8nec",
		"symphony162lkc9yrcf7j4dvlwrff4txyaf3acf5fhmesu4",
		"symphony1mfsymudfsdasnux2zmke5whthzt4rgmk906ah5",
		"symphony1fwju0v6rcwp6l94k30yshxsujlffgcatjp3fkk",
		"symphony1eumpa745safs0m0z6378c6fcapwr20p56n99qg",
		"symphony1h7c0s9sjvvhrghe8r3x47ukra6su2q9gvpc7up",
		"symphony1g0mm04cyw2sm5zh4x9r3y06ngusec6tmwjwsfd",
		"symphony1ptte2302zcukq9vcuwz2z6978cnx9lypv2sl58",
		"symphony1jhxx388tyz58mh8tqrfcgdpnaaj88y9l6dtmmc",
	}

	for _, acc := range airDropAccounts {
		genesisBalances[acc] = sdk.NewCoins(sdk.NewCoin("note", osmomath.NewInt(appParams.MicroUnit))) // 1 MLD
	}
}
