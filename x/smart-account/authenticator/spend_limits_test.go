package authenticator_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v27/x/txfees/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/ante"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/post"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/testutils"

	storetypes "cosmossdk.io/store/types"
)

type SpendLimitAuthenticatorTest struct {
	BaseAuthenticatorSuite

	Store                      prefix.Store
	CosmwasmAuth               authenticator.CosmwasmAuthenticator
	AlwaysPassAuth             testutils.TestingAuthenticator
	AuthenticatorAnteDecorator ante.AuthenticatorDecorator
	AuthenticatorPostDecorator post.AuthenticatorPostDecorator
}

type InstantiateMsg struct {
	PriceResolutionConfig PriceResolutionConfig `json:"price_resolution_config"`
	TrackedDenoms         []TrackedDenom        `json:"tracked_denoms"`
}

type TrackedDenom struct {
	Denom      string              `json:"denom"`
	SwapRoutes []SwapAmountInRoute `json:"swap_routes"`
}

type SwapAmountInRoute struct {
	PoolID        string `json:"pool_id"` // as u64
	TokenOutDenom string `json:"token_out_denom"`
}

type PriceResolutionConfig struct {
	QuoteDenom         string `json:"quote_denom"`
	StalenessThreshold string `json:"staleness_threshold"` // as u64
	TwapDuration       string `json:"twap_duration"`       // as u64
}

// params
type SpendLimitParams struct {
	Limit       string     `json:"limit"`        // as u128
	ResetPeriod string     `json:"reset_period"` // day | week | month | year
	TimeLimit   *TimeLimit `json:"time_limit,omitempty"`
}

type TimeLimit struct {
	Start *string `json:"start,omitempty"` // as u64 or None
	End   string  `json:"end"`             // as u64
}

func TestSpendLimitAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(SpendLimitAuthenticatorTest))
}

const UUSDC = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

func (s *SpendLimitAuthenticatorTest) SetupTest() {
	s.SetupKeys()

	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContextLegacy(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(10_000_000))
	s.Ctx = s.Ctx.WithBlockTime(time.Now())
	s.EncodingConfig = app.MakeEncodingConfig()

	s.CosmwasmAuth = authenticator.NewCosmwasmAuthenticator(s.OsmosisApp.ContractKeeper, s.OsmosisApp.AccountKeeper, s.OsmosisApp.AppCodec())

	s.AlwaysPassAuth = testutils.TestingAuthenticator{Approve: testutils.Always, Confirm: testutils.Always, GasConsumption: 0}
	s.OsmosisApp.SmartAccountKeeper.AuthenticatorManager.RegisterAuthenticator(s.AlwaysPassAuth)

	deductFeeDecorator := txfeeskeeper.NewDeductFeeDecorator(*s.OsmosisApp.TxFeesKeeper, s.OsmosisApp.AccountKeeper, s.OsmosisApp.BankKeeper, nil)
	s.AuthenticatorAnteDecorator = ante.NewAuthenticatorDecorator(
		s.OsmosisApp.AppCodec(),
		s.OsmosisApp.SmartAccountKeeper,
		s.OsmosisApp.AccountKeeper,
		s.EncodingConfig.TxConfig.SignModeHandler(),
		deductFeeDecorator,
	)

	s.AuthenticatorPostDecorator = post.NewAuthenticatorPostDecorator(
		s.OsmosisApp.AppCodec(),
		s.OsmosisApp.SmartAccountKeeper,
		s.OsmosisApp.AccountKeeper,
		s.EncodingConfig.TxConfig.SignModeHandler(),
		// Add an empty handler here to enable a circuit breaker pattern
		sdk.ChainPostDecorators(sdk.Terminator{}), //nolint
	)
}

func (s *SpendLimitAuthenticatorTest) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

func (s *SpendLimitAuthenticatorTest) TestSpendLimit() {
	anteHandler := sdk.ChainAnteDecorators(s.AuthenticatorAnteDecorator)
	postHandler := sdk.ChainPostDecorators(s.AuthenticatorPostDecorator)

	usdcOsmoPoolId := s.preparePool(
		[]balancer.PoolAsset{
			{
				Weight: osmomath.NewInt(100000),
				Token:  sdk.NewCoin(UUSDC, osmomath.NewInt(1500000000)),
			},
			{
				Weight: osmomath.NewInt(100000),
				Token:  sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000)),
			},
		},
	)

	// always update file name to reflect the version
	// current: https://github.com/osmosis-labs/spend-limit-authenticator/tree/1.0.0-alpha.1
	// most test cases exists in the repo above, this test file is intended to ensure that latest osmosis code
	// does not break existing contract
	codeId := s.StoreContractCode("../testutils/bytecode/spend_limit_v1.0.0-alpha.1.wasm")

	msg := InstantiateMsg{
		PriceResolutionConfig: PriceResolutionConfig{
			QuoteDenom:         UUSDC,
			StalenessThreshold: "3600000000000",
			TwapDuration:       "3600000000000",
		},
		TrackedDenoms: []TrackedDenom{
			{
				Denom: appparams.BaseCoinUnit,
				SwapRoutes: []SwapAmountInRoute{
					{
						PoolID:        fmt.Sprintf("%d", usdcOsmoPoolId),
						TokenOutDenom: UUSDC,
					},
				},
			},
		},
	}

	// increase time by 1hr to ensure twap price is available
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour))

	bz, err := json.Marshal(msg)
	s.Require().NoError(err)
	contractAddr := s.InstantiateContract(string(bz), codeId)

	// add new authenticator
	ak := s.OsmosisApp.AppKeepers.SmartAccountKeeper

	authAcc := s.TestAccAddress[1]
	authAccPriv := s.TestPrivKeys[1]

	params := SpendLimitParams{
		Limit:       "5000000000",
		ResetPeriod: "day",
		TimeLimit: &TimeLimit{
			Start: nil,
			End:   fmt.Sprintf("%d", time.Now().Add(time.Hour*25).UnixNano()),
		},
	}

	bz, err = json.Marshal(params)

	s.Require().NoError(err)

	initData := authenticator.CosmwasmAuthenticatorInitData{
		Contract: contractAddr.String(),
		Params:   bz,
	}

	bz, err = json.Marshal(initData)
	s.Require().NoError(err)

	// hack to get fee payer authenticated
	id, err := ak.AddAuthenticator(s.Ctx, authAcc, s.AlwaysPassAuth.Type(), []byte{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), id)

	id, err = ak.AddAuthenticator(s.Ctx, authAcc, authenticator.CosmwasmAuthenticator{}.Type(), bz)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), id)

	// fund acc
	s.FundAcc(authAcc, sdk.NewCoins(sdk.NewCoin(UUSDC, osmomath.NewInt(200000000000))))
	s.FundAcc(authAcc, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(200000000000))))

	// a hack for setting fee payer
	selfSend := banktypes.MsgSend{
		FromAddress: authAcc.String(),
		ToAddress:   authAcc.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(UUSDC, osmomath.NewInt(1))),
	}

	// swap within limit
	swapMsg := poolmanagertypes.MsgSwapExactAmountIn{
		Sender:            authAcc.String(),
		Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: usdcOsmoPoolId, TokenOutDenom: UUSDC}},
		TokenIn:           sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(3333333333)), // ~ 4,999,999,999 uusdc
		TokenOutMinAmount: osmomath.OneInt(),
	}

	tx, err := s.GenSimpleTxWithSelectedAuthenticators([]sdk.Msg{&selfSend, &swapMsg}, []cryptotypes.PrivKey{authAccPriv}, []uint64{1, 2})
	s.Require().NoError(err)

	// ante

	_, err = anteHandler(s.Ctx, tx, false)
	s.Require().NoError(err)

	// swap
	_, err = s.OsmosisApp.MsgServiceRouter().Handler(&swapMsg)(s.Ctx, &swapMsg)
	s.Require().NoError(err)

	// post

	_, err = postHandler(s.Ctx, tx, false, true)
	s.Require().NoError(err)

	// swap over the limit
	swapMsg = poolmanagertypes.MsgSwapExactAmountIn{
		Sender:            authAcc.String(),
		Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: usdcOsmoPoolId, TokenOutDenom: appparams.BaseCoinUnit}},
		TokenIn:           sdk.NewCoin(UUSDC, osmomath.NewInt(2)),
		TokenOutMinAmount: osmomath.OneInt(),
	}

	tx, err = s.GenSimpleTxWithSelectedAuthenticators([]sdk.Msg{&selfSend, &swapMsg}, []cryptotypes.PrivKey{authAccPriv}, []uint64{1, 2})
	s.Require().NoError(err)

	// ante
	_, err = anteHandler(s.Ctx, tx, false)
	s.Require().NoError(err)

	// swap
	_, err = s.OsmosisApp.MsgServiceRouter().Handler(&swapMsg)(s.Ctx, &swapMsg)
	s.Require().NoError(err)

	// post
	_, err = postHandler(s.Ctx, tx, false, true)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "Spend limit error: Overspend: remaining qouta 1, requested 2: execute wasm contract failed")

	// advance time to next day, and resend the prev tx should have no error
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

	// ante
	_, err = anteHandler(s.Ctx, tx, false)
	s.Require().NoError(err)

	// swap
	_, err = s.OsmosisApp.MsgServiceRouter().Handler(&swapMsg)(s.Ctx, &swapMsg)
	s.Require().NoError(err)

	// post
	_, err = postHandler(s.Ctx, tx, false, true)
	s.Require().NoError(err)

	// advance time to end time, that should fail
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour))

	// ante
	_, err = anteHandler(s.Ctx, tx, false)
	s.Require().Error(err)

	nanoDigits := 9
	endTimeSecsStr := params.TimeLimit.End[:len(params.TimeLimit.End)-nanoDigits]
	endTimeNanosStr := params.TimeLimit.End[len(params.TimeLimit.End)-nanoDigits:]

	s.Require().Contains(
		err.Error(),
		fmt.Sprintf(
			"Current time %d.%09d not within time limit None - %s.%s: execute wasm contract failed",
			s.Ctx.BlockTime().Unix(), s.Ctx.BlockTime().Nanosecond(),
			endTimeSecsStr, endTimeNanosStr,
		),
	)
}

func (s *SpendLimitAuthenticatorTest) StoreContractCode(path string) uint64 {
	osmosisApp := s.OsmosisApp
	govKeeper := wasmkeeper.NewGovPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	wasmCode, err := os.ReadFile(path)
	s.Require().NoError(err)
	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(s.Ctx, creator, wasmCode, &accessEveryone)
	s.Require().NoError(err)
	return codeID
}

func (s *SpendLimitAuthenticatorTest) InstantiateContract(msg string, codeID uint64) sdk.AccAddress {
	osmosisApp := s.OsmosisApp
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(s.Ctx, codeID, creator, creator, []byte(msg), "contract", nil)
	s.Require().NoError(err)
	return addr
}

func (s *SpendLimitAuthenticatorTest) preparePool(
	poolAssets []balancer.PoolAsset,
) uint64 {
	poolCreator := s.TestAccAddress[0]

	s.FundAcc(poolCreator, s.OsmosisApp.PoolManagerKeeper.GetParams(s.Ctx).PoolCreationFee)

	for _, asset := range poolAssets {
		s.FundAcc(poolCreator, sdk.NewCoins(asset.Token))
	}

	poolParams := balancer.PoolParams{
		SwapFee: osmomath.ZeroDec(),
		ExitFee: osmomath.ZeroDec(),
	}

	poolID, err := s.OsmosisApp.PoolManagerKeeper.CreatePool(
		s.Ctx,
		balancer.NewMsgCreateBalancerPool(poolCreator, poolParams, poolAssets, ""),
	)
	s.Require().NoError(err)

	return poolID
}
