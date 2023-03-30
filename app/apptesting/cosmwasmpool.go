package apptesting

import (
	"encoding/json"
	"os"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/model"

	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
)

const (
	DefaultTransmuterDenomA = "axlusdc"
	DefaultTransmuterDenomB = "gravusdc"
	DefaultCodeId           = 1
)

// PrepareCosmWasmPool sets up a cosmwasm pool with the default parameters.
func (s *KeeperTestHelper) PrepareCosmWasmPool() cosmwasmpooltypes.CosmWasmExtension {
	return s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{DefaultTransmuterDenomA, DefaultTransmuterDenomB}, DefaultCodeId)
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomTransmuterPool(owner sdk.AccAddress, denoms []string, codeId uint64) cosmwasmpooltypes.CosmWasmExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	cosmwasmpoolModuleAddr := s.App.AccountKeeper.GetModuleAddress(cosmwasmpooltypes.ModuleName)
	s.Require().NotNil(cosmwasmpoolModuleAddr)

	// TODO: make sure permissiosns are updated in the upgrade handler.
	s.App.WasmKeeper.SetParams(s.Ctx, wasmtypes.Params{
		CodeUploadAccess: wasmtypes.AccessConfig{
			Permission: wasmtypes.AccessTypeAnyOfAddresses,
			Addresses:  []string{cosmwasmpoolModuleAddr.String()},
		},
		InstantiateDefaultPermission: wasmtypes.AccessTypeAnyOfAddresses,
	})

	workingDir, err := os.Getwd()
	s.Require().NoError(err)

	projectRootPath := "/osmosis/"
	projectRootIndex := strings.LastIndex(workingDir, projectRootPath) + len(projectRootPath)
	workingDir = workingDir[:projectRootIndex]

	code, err := os.ReadFile(workingDir + "x/cosmwasmpool/bytecode/transmuter.wasm")
	s.Require().NoError(err)

	s.Require().NoError(err)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: cosmwasmpoolModuleAddr.String()}
	codeID, _, err := s.App.ContractKeeper.Create(s.Ctx, cosmwasmpoolModuleAddr, code, &instantiateConfig)
	s.Require().NoError(err)

	instantiateMsg := msg.InstantiateMsg{
		PoolAssetDenoms: denoms,
	}

	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)

	addr, _, err := s.App.ContractKeeper.Instantiate(s.Ctx, codeID, cosmwasmpoolModuleAddr, cosmwasmpoolModuleAddr, instantiateMsgBz, "transmuter contract", nil)
	s.Require().NoError(err)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	s.App.PoolManagerKeeper.SetNextPoolId(s.Ctx, nextPoolId+1)

	pool := model.NewCosmWasmPool(nextPoolId, codeID, []byte{})
	pool.SetContractAddress(addr.String())
	pool.SetWasmKeeper(s.App.WasmKeeper)

	s.App.CosmwasmPoolKeeper.SetPool(s.Ctx, &pool)

	return &pool
}
