package keeper

////nolint
////DONTCOVER
//
//import (
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/require"
//
//	customauth "github.com/classic-terra/core/v2/custom/auth"
//	custombank "github.com/classic-terra/core/v2/custom/bank"
//	customdistr "github.com/classic-terra/core/v2/custom/distribution"
//	customparams "github.com/classic-terra/core/v2/custom/params"
//	customstaking "github.com/classic-terra/core/v2/custom/staking"
//	core "github.com/classic-terra/core/v2/types"
//	"github.com/classic-terra/core/v2/x/market/types"
//	"github.com/classic-terra/core/v2/x/oracle"
//	oraclekeeper "github.com/classic-terra/core/v2/x/oracle/keeper"
//	oracletypes "github.com/classic-terra/core/v2/x/oracle/types"
//
//	"github.com/tendermint/tendermint/crypto"
//	"github.com/tendermint/tendermint/crypto/secp256k1"
//	"github.com/tendermint/tendermint/libs/log"
//	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
//	dbm "github.com/tendermint/tm-db"
//
//	simparams "cosmossdk.io/simapp/params"
//	"github.com/cosmos/cosmos-sdk/codec"
//	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
//	"github.com/cosmos/cosmos-sdk/std"
//	"github.com/cosmos/cosmos-sdk/store"
//	storetypes "github.com/cosmos/cosmos-sdk/store/types"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/module"
//	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
//	"github.com/cosmos/cosmos-sdk/x/auth/tx"
//	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
//	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
//	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
//	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
//	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
//	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
//	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
//	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
//	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
//)
//
//const faucetAccountName = "faucet"
//
//var ModuleBasics = module.NewBasicManager(
//	customauth.AppModuleBasic{},
//	custombank.AppModuleBasic{},
//	customstaking.AppModuleBasic{},
//	customdistr.AppModuleBasic{},
//	customparams.AppModuleBasic{},
//	oracle.AppModuleBasic{},
//)
//
//// MakeTestCodec
//func MakeTestCodec(t *testing.T) codec.Codec {
//	return MakeEncodingConfig(t).Codec
//}
//
//// MakeEncodingConfig
//func MakeEncodingConfig(_ *testing.T) simparams.EncodingConfig {
//	amino := codec.NewLegacyAmino()
//	interfaceRegistry := codectypes.NewInterfaceRegistry()
//	codec := codec.NewProtoCodec(interfaceRegistry)
//	txCfg := tx.NewTxConfig(codec, tx.DefaultSignModes)
//
//	std.RegisterInterfaces(interfaceRegistry)
//	std.RegisterLegacyAminoCodec(amino)
//
//	ModuleBasics.RegisterLegacyAminoCodec(amino)
//	ModuleBasics.RegisterInterfaces(interfaceRegistry)
//	types.RegisterLegacyAminoCodec(amino)
//	types.RegisterInterfaces(interfaceRegistry)
//
//	return simparams.EncodingConfig{
//		InterfaceRegistry: interfaceRegistry,
//		Codec:             codec,
//		TxConfig:          txCfg,
//		Amino:             amino,
//	}
//}
//
//// Test Account
//var (
//	PubKeys = []crypto.PubKey{
//		secp256k1.GenPrivKey().PubKey(),
//		secp256k1.GenPrivKey().PubKey(),
//		secp256k1.GenPrivKey().PubKey(),
//	}
//
//	Addrs = []sdk.AccAddress{
//		sdk.AccAddress(PubKeys[0].Address()),
//		sdk.AccAddress(PubKeys[1].Address()),
//		sdk.AccAddress(PubKeys[2].Address()),
//	}
//
//	ValAddrs = []sdk.ValAddress{
//		sdk.ValAddress(PubKeys[0].Address()),
//		sdk.ValAddress(PubKeys[1].Address()),
//		sdk.ValAddress(PubKeys[2].Address()),
//	}
//
//	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
//	InitCoins  = sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, InitTokens))
//)
//
//type TestInput struct {
//	Ctx           sdk.Context
//	Cdc           *codec.LegacyAmino
//	AccountKeeper authkeeper.AccountKeeper
//	BankKeeper    bankkeeper.Keeper
//	OracleKeeper  types.OracleKeeper
//	MarketKeeper  Keeper
//}
//
//func CreateTestInput(t *testing.T) TestInput {
//	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
//	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
//	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
//	tKeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
//	keyOracle := sdk.NewKVStoreKey(oracletypes.StoreKey)
//	keyStaking := sdk.NewKVStoreKey(stakingtypes.StoreKey)
//	keyDistr := sdk.NewKVStoreKey(distrtypes.StoreKey)
//	keyMarket := sdk.NewKVStoreKey(types.StoreKey)
//
//	db := dbm.NewMemDB()
//	ms := store.NewCommitMultiStore(db)
//	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())
//	encodingConfig := MakeEncodingConfig(t)
//	appCodec, legacyAmino := encodingConfig.Codec, encodingConfig.Amino
//
//	ms.MountStoreWithDB(keyAcc, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(keyBank, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(tKeyParams, storetypes.StoreTypeTransient, db)
//	ms.MountStoreWithDB(keyParams, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(keyOracle, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(keyStaking, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(keyDistr, storetypes.StoreTypeIAVL, db)
//	ms.MountStoreWithDB(keyMarket, storetypes.StoreTypeIAVL, db)
//
//	require.NoError(t, ms.LoadLatestVersion())
//
//	blackListAddrs := map[string]bool{
//		faucetAccountName:              true,
//		authtypes.FeeCollectorName:     true,
//		stakingtypes.NotBondedPoolName: true,
//		stakingtypes.BondedPoolName:    true,
//		distrtypes.ModuleName:          true,
//	}
//
//	maccPerms := map[string][]string{
//		faucetAccountName:              {authtypes.Minter},
//		authtypes.FeeCollectorName:     nil,
//		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
//		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
//		distrtypes.ModuleName:          nil,
//		oracletypes.ModuleName:         nil,
//		types.ModuleName:               {authtypes.Burner, authtypes.Minter},
//	}
//
//	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, keyParams, tKeyParams)
//	accountKeeper := authkeeper.NewAccountKeeper(appCodec, keyAcc, paramsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms, sdk.GetConfig().GetBech32AccountAddrPrefix())
//	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, keyBank, accountKeeper, paramsKeeper.Subspace(banktypes.ModuleName), blackListAddrs)
//
//	totalSupply := sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, InitTokens.MulRaw(int64(len(Addrs)*10))))
//	err := bankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)
//	require.NoError(t, err)
//
//	stakingKeeper := stakingkeeper.NewKeeper(
//		appCodec,
//		keyStaking,
//		accountKeeper,
//		bankKeeper,
//		paramsKeeper.Subspace(stakingtypes.ModuleName),
//	)
//
//	stakingParams := stakingtypes.DefaultParams()
//	stakingParams.BondDenom = core.MicroLunaDenom
//	stakingKeeper.SetParams(ctx, stakingParams)
//
//	distrKeeper := distrkeeper.NewKeeper(
//		appCodec,
//		keyDistr, paramsKeeper.Subspace(distrtypes.ModuleName),
//		accountKeeper, bankKeeper, &stakingKeeper,
//		authtypes.FeeCollectorName)
//
//	distrKeeper.SetFeePool(ctx, distrtypes.InitialFeePool())
//	distrParams := distrtypes.DefaultParams()
//	distrParams.CommunityTax = sdk.NewDecWithPrec(2, 2)
//	distrParams.BaseProposerReward = sdk.NewDecWithPrec(1, 2)
//	distrParams.BonusProposerReward = sdk.NewDecWithPrec(4, 2)
//	distrKeeper.SetParams(ctx, distrParams)
//	stakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(distrKeeper.Hooks()))
//
//	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
//	notBondedPool := authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
//	bondPool := authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName, authtypes.Burner, authtypes.Staking)
//	distrAcc := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)
//	oracleAcc := authtypes.NewEmptyModuleAccount(oracletypes.ModuleName)
//	marketAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Burner, authtypes.Minter)
//
//	err = bankKeeper.SendCoinsFromModuleToModule(ctx, faucetAccountName, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, InitTokens.MulRaw(int64(len(Addrs))))))
//	require.NoError(t, err)
//
//	accountKeeper.SetModuleAccount(ctx, feeCollectorAcc)
//	accountKeeper.SetModuleAccount(ctx, bondPool)
//	accountKeeper.SetModuleAccount(ctx, notBondedPool)
//	accountKeeper.SetModuleAccount(ctx, distrAcc)
//	accountKeeper.SetModuleAccount(ctx, oracleAcc)
//	accountKeeper.SetModuleAccount(ctx, marketAcc)
//
//	for _, addr := range Addrs {
//		accountKeeper.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(addr))
//		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, faucetAccountName, addr, InitCoins)
//		require.NoError(t, err)
//		require.Equal(t, bankKeeper.GetAllBalances(ctx, addr), InitCoins)
//	}
//
//	oracleKeeper := oraclekeeper.NewKeeper(
//		appCodec,
//		keyOracle,
//		paramsKeeper.Subspace(oracletypes.ModuleName),
//		accountKeeper,
//		bankKeeper,
//		distrKeeper,
//		stakingKeeper,
//		distrtypes.ModuleName,
//	)
//	oracleDefaultParams := oracletypes.DefaultParams()
//	oracleKeeper.SetParams(ctx, oracleDefaultParams)
//
//	for _, denom := range oracleDefaultParams.Whitelist {
//		oracleKeeper.SetTobinTax(ctx, denom.Name, denom.TobinTax)
//	}
//
//	keeper := NewKeeper(
//		appCodec,
//		keyMarket, paramsKeeper.Subspace(types.ModuleName),
//		accountKeeper,
//		bankKeeper,
//		oracleKeeper,
//	)
//	keeper.SetParams(ctx, types.DefaultParams())
//
//	return TestInput{ctx, legacyAmino, accountKeeper, bankKeeper, oracleKeeper, keeper}
//}
//
//// FundAccount is a utility function that funds an account by minting and
//// sending the coins to the address. This should be used for testing purposes
//// only!
//func FundAccount(input TestInput, addr sdk.AccAddress, amounts sdk.Coins) error {
//	if err := input.BankKeeper.MintCoins(input.Ctx, faucetAccountName, amounts); err != nil {
//		return err
//	}
//
//	return input.BankKeeper.SendCoinsFromModuleToAccount(input.Ctx, faucetAccountName, addr, amounts)
//}
