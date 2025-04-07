package keeper

import (
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/noapptest"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	epochskeeper "github.com/osmosis-labs/osmosis/v27/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/market"
	marketkeeper "github.com/osmosis-labs/osmosis/v27/x/market/keeper"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"github.com/osmosis-labs/osmosis/v27/x/oracle"
	oraclekeeper "github.com/osmosis-labs/osmosis/v27/x/oracle/keeper"
	oracletypes "github.com/osmosis-labs/osmosis/v27/x/oracle/types"
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const faucetAccountName = "faucet"

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	distr.AppModuleBasic{},
	params.AppModuleBasic{},
	oracle.AppModuleBasic{},
	market.AppModuleBasic{},
)

// MakeTestCodec
func MakeTestCodec(t *testing.T) codec.Codec {
	return MakeEncodingConfig(t).Marshaler
}

// MakeEncodingConfig
func MakeEncodingConfig(_ *testing.T) appparams.EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(codec, tx.DefaultSignModes)

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	ModuleBasics.RegisterLegacyAminoCodec(amino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	return appparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         codec,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

// Test Account
var (
	PubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	Addrs = []sdk.AccAddress{
		sdk.AccAddress(PubKeys[0].Address()),
		sdk.AccAddress(PubKeys[1].Address()),
		sdk.AccAddress(PubKeys[2].Address()),
	}

	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(PubKeys[0].Address()),
		sdk.ValAddress(PubKeys[1].Address()),
		sdk.ValAddress(PubKeys[2].Address()),
	}

	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitCoins  = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens))
)

type TestInput struct {
	Ctx            sdk.Context
	Cdc            *codec.LegacyAmino
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	OracleKeeper   types.OracleKeeper
	MarketKeeper   marketkeeper.Keeper
	TreasuryKeeper Keeper
}

func CreateTestInput(t *testing.T) TestInput {
	t.Helper()
	keyAcc := storetypes.NewKVStoreKey(authtypes.StoreKey)
	keyBank := storetypes.NewKVStoreKey(banktypes.StoreKey)
	keyParams := storetypes.NewKVStoreKey(paramstypes.StoreKey)
	tKeyParams := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	keyOracle := storetypes.NewKVStoreKey(oracletypes.StoreKey)
	keyStaking := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	keyDistr := storetypes.NewKVStoreKey(distrtypes.StoreKey)
	keyMarket := storetypes.NewKVStoreKey(types.StoreKey)
	keyEpochs := storetypes.NewKVStoreKey(epochstypes.StoreKey)

	encodingConfig := MakeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	ctx := noapptest.CtxWithStoreKeys([]storetypes.StoreKey{
		keyAcc,
		keyBank,
		tKeyParams,
		keyParams,
		keyOracle,
		keyStaking,
		keyDistr,
		keyMarket,
	}, tmproto.Header{Time: time.Now().UTC()}, false)

	blackListAddrs := map[string]bool{
		faucetAccountName:              true,
		authtypes.FeeCollectorName:     true,
		stakingtypes.NotBondedPoolName: true,
		stakingtypes.BondedPoolName:    true,
		distrtypes.ModuleName:          true,
	}

	maccPerms := map[string][]string{
		faucetAccountName:              {authtypes.Minter},
		authtypes.FeeCollectorName:     nil,
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		distrtypes.ModuleName:          nil,
		oracletypes.ModuleName:         nil,
		markettypes.ModuleName:         {authtypes.Burner, authtypes.Minter},
		types.ModuleName:               nil,
	}

	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, keyParams, tKeyParams)
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keyAcc),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		"melody",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keyBank),
		accountKeeper,
		blackListAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)

	epochKeeper := epochskeeper.NewKeeper(keyEpochs)

	var err error

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keyStaking),
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = appparams.BaseCoinUnit
	err = stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	distrKeeper := distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keyDistr),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	err = distrKeeper.FeePool.Set(ctx, distrtypes.InitialFeePool())
	require.NoError(t, err)
	distrParams := distrtypes.DefaultParams()
	distrParams.CommunityTax = osmomath.NewDecWithPrec(2, 2)
	distrParams.BaseProposerReward = osmomath.NewDecWithPrec(1, 2)
	distrParams.BonusProposerReward = osmomath.NewDecWithPrec(4, 2)
	err = distrKeeper.Params.Set(ctx, distrParams)
	require.NoError(t, err)
	stakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(distrKeeper.Hooks()))

	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	notBondedPool := authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	bondPool := authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName, authtypes.Burner, authtypes.Staking)
	distrAcc := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)
	oracleAcc := authtypes.NewEmptyModuleAccount(oracletypes.ModuleName)
	marketAcc := authtypes.NewEmptyModuleAccount(markettypes.ModuleName, authtypes.Burner, authtypes.Minter)
	treasuryAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	faucetAcc := authtypes.NewEmptyModuleAccount(faucetAccountName, authtypes.Minter)

	for index, acc := range []*authtypes.ModuleAccount{
		faucetAcc,
		feeCollectorAcc,
		bondPool,
		notBondedPool,
		distrAcc,
		oracleAcc,
		marketAcc,
		treasuryAcc,
	} {
		acc.AccountNumber = uint64(index)
		accountKeeper.SetModuleAccount(ctx, acc)
	}

	totalSupply := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addrs)*10))))
	err = bankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)
	require.NoError(t, err)

	// mint stable
	totalSupply = sdk.NewCoins(sdk.NewCoin(assets.MicroSDRDenom, sdkmath.NewInt(10_000*1e6)))
	err = bankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)
	require.NoError(t, err)

	err = bankKeeper.SendCoinsFromModuleToModule(ctx, faucetAccountName, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addrs))))))
	require.NoError(t, err)

	for index, addr := range Addrs {
		acc := authtypes.NewBaseAccountWithAddress(addr)
		acc.AccountNumber = uint64(index + 1000)
		accountKeeper.SetAccount(ctx, acc)
		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, faucetAccountName, addr, InitCoins)
		require.NoError(t, err)
		require.Equal(t, bankKeeper.GetAllBalances(ctx, addr), InitCoins)
	}

	oracleKeeper := oraclekeeper.NewKeeper(
		appCodec,
		keyOracle,
		paramsKeeper.Subspace(oracletypes.ModuleName),
		accountKeeper,
		bankKeeper,
		distrKeeper,
		stakingKeeper,
		epochKeeper,
		distrtypes.ModuleName,
	)
	oracleDefaultParams := oracletypes.DefaultParams()
	oracleDefaultParams.Whitelist = oracletypes.DenomList{oracletypes.Denom{
		Name:     assets.MicroSDRDenom,
		TobinTax: osmomath.ZeroDec(),
	}}
	oracleKeeper.SetParams(ctx, oracleDefaultParams)
	oracleKeeper.SetMelodyExchangeRate(ctx, assets.MicroSDRDenom, osmomath.NewDecWithPrec(1, 1))

	for _, denom := range oracleDefaultParams.Whitelist {
		oracleKeeper.SetTobinTax(ctx, denom.Name, denom.TobinTax)
	}

	marketKeeper := marketkeeper.NewKeeper(appCodec,
		keyMarket, paramsKeeper.Subspace(markettypes.ModuleName),
		accountKeeper,
		bankKeeper,
		oracleKeeper)

	keeper := NewKeeper(
		appCodec,
		keyMarket,
		paramsKeeper.Subspace(types.ModuleName),
		accountKeeper,
		bankKeeper,
		marketKeeper,
		oracleKeeper,
	)
	keeper.SetParams(ctx, types.DefaultParams())

	return TestInput{
		Ctx:            ctx,
		Cdc:            legacyAmino,
		AccountKeeper:  accountKeeper,
		BankKeeper:     bankKeeper,
		OracleKeeper:   oracleKeeper,
		MarketKeeper:   marketKeeper,
		TreasuryKeeper: keeper,
	}
}

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address. This should be used for testing purposes
// only!
func FundAccount(input TestInput, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := input.BankKeeper.MintCoins(input.Ctx, faucetAccountName, amounts); err != nil {
		return err
	}

	return input.BankKeeper.SendCoinsFromModuleToAccount(input.Ctx, faucetAccountName, addr, amounts)
}
