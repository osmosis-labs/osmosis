package authenticator_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	minttypes "github.com/osmosis-labs/osmosis/v23/x/mint/types"

	"github.com/osmosis-labs/osmosis/v23/app"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
)

type SpendLimitAuthenticatorTest struct {
	BaseAuthenticatorSuite

	Store        prefix.Store
	CosmwasmAuth authenticator.CosmwasmAuthenticator
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
	PoolID        uint64 `json:"pool_id"`
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

func (s *SpendLimitAuthenticatorTest) SetupTest() {
	s.SetupKeys()

	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(10_000_000))
	s.Ctx = s.Ctx.WithBlockTime(time.Now())
	s.EncodingConfig = app.MakeEncodingConfig()

	s.CosmwasmAuth = authenticator.NewCosmwasmAuthenticator(s.OsmosisApp.ContractKeeper, s.OsmosisApp.AccountKeeper, s.EncodingConfig.TxConfig.SignModeHandler(), s.OsmosisApp.AppCodec())

	amount := 1000000000
	coins := sdk.NewCoins(sdk.NewInt64Coin("uosmo", int64(amount)))
	err := s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, coins)
	s.Require().NoError(err, "Failed mint coins")

}

func (s *SpendLimitAuthenticatorTest) TestSpendLimit() {
	// always update file name to reflect the version
	// current: https://github.com/osmosis-labs/spend-limit-authenticator/tree/1.0.0-alpha.1
	// most test cases exists the repo above, this test file is intended to ensure that latest osmosis code
	// does not break existing contract
	codeId := s.StoreContractCode("../testutils/bytecode/spend_limit_v1.0.0-alpha.1.wasm")

	msg := InstantiateMsg{
		PriceResolutionConfig: PriceResolutionConfig{
			QuoteDenom:         "uosmo",
			StalenessThreshold: "1000",
			TwapDuration:       "1000",
		},
		TrackedDenoms: []TrackedDenom{},
	}

	bz, err := json.Marshal(msg)
	s.Require().NoError(err)
	contractAddr := s.InstantiateContract(string(bz), codeId)

	// add new authenticator
	ak := s.OsmosisApp.AppKeepers.AuthenticatorKeeper

	acc := s.TestAccAddress[0]

	params := SpendLimitParams{
		Limit:       "1000000000",
		ResetPeriod: "day",
	}
	bz, err = json.Marshal(params)

	s.Require().NoError(err)

	initData := authenticator.CosmwasmAuthenticatorInitData{
		Contract: contractAddr.String(),
		Params:   bz,
	}

	bz, err = json.Marshal(initData)
	s.Require().NoError(err)

	id, err := ak.AddAuthenticator(s.Ctx, acc, authenticator.CosmwasmAuthenticator{}.Type(), bz)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), id)
}

func (s *SpendLimitAuthenticatorTest) StoreContractCode(path string) uint64 {
	osmosisApp := s.OsmosisApp
	govKeeper := wasmkeeper.NewGovPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	wasmCode, err := os.ReadFile(path)
	s.Require().NoError(err)
	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(s.Ctx.WithBlockTime(time.Now()), creator, wasmCode, &accessEveryone)
	s.Require().NoError(err)
	return codeID
}

func (s *SpendLimitAuthenticatorTest) InstantiateContract(msg string, codeID uint64) sdk.AccAddress {
	osmosisApp := s.OsmosisApp
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(s.Ctx.WithBlockTime(time.Now()), codeID, creator, creator, []byte(msg), "contract", nil)
	s.Require().NoError(err)
	return addr
}

// func (s *CosmwasmAuthenticatorTest) QueryContract(msg string, contractAddr sdk.AccAddress) []byte {
// 	// Query the contract
// 	osmosisApp := s.OsmosisApp
// 	res, err := osmosisApp.WasmKeeper.QuerySmart(s.Ctx.WithBlockTime(time.Now()), contractAddr, []byte(msg))
// 	s.Require().NoError(err)

// 	return res
// }
