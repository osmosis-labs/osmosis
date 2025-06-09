package app

import (
	"encoding/json"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"os"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cosmosdb "github.com/cosmos/cosmos-db"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdkmath "cosmossdk.io/math"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func GenesisStateWithValSet(app *SymphonyApp) GenesisState {
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	senderPrivKey.PubKey().Address()
	acc := authtypes.NewBaseAccountWithAddress(senderPrivKey.PubKey().Address().Bytes())

	//////////////////////
	balances := []banktypes.Balance{}
	genesisState := NewDefaultGenesisState()
	genAccs := []authtypes.GenesisAccount{acc}
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction
	initValPowers := []abci.ValidatorUpdate{}

	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   osmomath.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), osmomath.OneDec()))

		// add initial validator powers so consumer InitGenesis runs correctly
		pub, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: pub.PubKey,
		})
	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	_, err := tmtypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	// set validators and delegations
	marketGenesis := markettypes.NewGenesisState(markettypes.DefaultParams())
	taxReceiver, err := sdk.AccAddressFromBech32("symphony1h0jhfjfpqc8463fd040w0wxme9aq5ujtclf3np")
	marketGenesis.Params.TaxReceiver = taxReceiver.String()
	genesisState[markettypes.ModuleName] = app.AppCodec().MustMarshalJSON(marketGenesis)

	return genesisState
}

var defaultGenesisStatebytes = []byte{}

// SetupWithCustomHome initializes a new SymphonyApp with a custom home directory
func SetupWithCustomHome(isCheckTx bool, dir string) *SymphonyApp {
	return SetupWithCustomHomeAndChainId(isCheckTx, dir, "osmosis-1")
}

func SetupWithCustomHomeAndChainId(isCheckTx bool, dir, chainId string) *SymphonyApp {
	db := cosmosdb.NewMemDB()
	app := NewSymphonyApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, dir, 0, sims.EmptyAppOptions{}, EmptyWasmOpts, baseapp.SetChainID(chainId))
	if !isCheckTx {
		if len(defaultGenesisStatebytes) == 0 {
			var err error
			genesisState := GenesisStateWithValSet(app)
			defaultGenesisStatebytes, err = json.Marshal(genesisState)
			if err != nil {
				panic(err)
			}
		}

		_, err := app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   defaultGenesisStatebytes,
				ChainId:         chainId,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// Setup initializes a new SymphonyApp.
func Setup(isCheckTx bool) *SymphonyApp {
	return SetupWithCustomHome(isCheckTx, DefaultNodeHome)
}

// SetupTestingAppWithLevelDb initializes a new SymphonyApp intended for testing,
// with LevelDB as a db.
func SetupTestingAppWithLevelDb(isCheckTx bool) (app *SymphonyApp, cleanupFn func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "osmosis_leveldb_testing")
	if err != nil {
		panic(err)
	}
	db, err := cosmosdb.NewGoLevelDB("osmosis_leveldb_testing", dir, nil)
	if err != nil {
		panic(err)
	}

	app = NewSymphonyApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, sims.EmptyAppOptions{}, EmptyWasmOpts, baseapp.SetChainID("osmosis-1"))
	if !isCheckTx {
		genesisState := GenesisStateWithValSet(app)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
				ChainId:         "osmosis-1",
			},
		)
		if err != nil {
			panic(err)
		}
	}

	cleanupFn = func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}

	return app, cleanupFn
}
