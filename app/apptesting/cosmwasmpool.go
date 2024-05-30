package apptesting

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
	"github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/cosmwasm/msg/transmuter"
	v3 "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/cosmwasm/msg/v3"
	"github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/model"

	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/types"
)

const (
	DefaultTransmuterDenomA       = "axlusdc"
	DefaultTransmuterDenomB       = "gravusdc"
	TransmuterContractName        = "transmuter"
	TransmuterV3ContractName      = "transmuterv3"
	TransmuterMigrateContractName = "transmuter_migrate"
	DefaultCodeId                 = 1

	osmosisRepository         = "osmosis"
	osmosisRepoTransmuterPath = "x/cosmwasmpool/bytecode"
)

// PrepareCosmWasmPool sets up a cosmwasm pool with the default parameters.
func (s *KeeperTestHelper) PrepareCosmWasmPool() cosmwasmpooltypes.CosmWasmExtension {
	return s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{DefaultTransmuterDenomA, DefaultTransmuterDenomB})
}

// PrepareCustomTransmuterPool sets up a transmuter pool with the default parameters assuming that
// the transmuter contract is stored under x/cosmwasmpool/bytecode in the Osmosis repository.
func (s *KeeperTestHelper) PrepareCustomTransmuterPool(owner sdk.AccAddress, denoms []string) cosmwasmpooltypes.CosmWasmExtension {
	return s.PrepareCustomTransmuterPoolCustomProject(owner, denoms, osmosisRepository, osmosisRepoTransmuterPath)
}

// PrepareCustomTransmuterPoolCustomProject sets up a transmuter pool with the custom parameters.
// Gives flexibility for the helper to be reused outside of the Osmosis repository by providing custom
// project name and bytecode path.
func (s *KeeperTestHelper) PrepareCustomTransmuterPoolCustomProject(owner sdk.AccAddress, denoms []string, projectName, byteCodePath string) cosmwasmpooltypes.CosmWasmExtension {
	return s.PreparePool(owner, denoms, nil, projectName, byteCodePath, TransmuterContractName, s.GetTransmuterInstantiateMsgBytes)
}

// PrepareCustomTransmuterPoolV3 sets up a transmuter pool with the custom parameters for version 3 of the transmuter contract (alloyed assets).
// It initializes the pool with the provided ratio for the given denoms, using a default normalization factor of "1" for each denom.
func (s *KeeperTestHelper) PrepareCustomTransmuterPoolV3(owner sdk.AccAddress, denoms []string, ratio []uint16) cosmwasmpooltypes.CosmWasmExtension {
	normalizationFactors := make([]string, len(denoms))
	for i := range normalizationFactors {
		normalizationFactors[i] = "1"
	}
	return s.PrepareCustomTransmuterPoolV3WithNormalization(owner, denoms, normalizationFactors, ratio)
}

// PrepareCustomTransmuterPoolV3WithNormalization sets up a transmuter pool with the custom parameters for version 3 of the transmuter contract (alloyed assets).
// It initializes the pool with the provided ratio for the given denoms and their respective normalization factors.
func (s *KeeperTestHelper) PrepareCustomTransmuterPoolV3WithNormalization(owner sdk.AccAddress, denoms []string, normalizationFactors []string, ratio []uint16) cosmwasmpooltypes.CosmWasmExtension {
	pool := s.PreparePool(owner, denoms, normalizationFactors, osmosisRepository, osmosisRepoTransmuterPath, TransmuterV3ContractName, s.GetTransmuterInstantiateMsgBytesV3)
	s.AddRatioFundsToTransmuterPool(s.TestAccs[0], denoms, ratio, pool.GetId())
	pool, err := s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)
	return pool
}

// PreparePool sets up a transmuter pool with the custom parameters and optional normalization factors.
func (s *KeeperTestHelper) PreparePool(owner sdk.AccAddress, denoms []string, normalizationFactors []string, projectName, byteCodePath, contractName string, getInstantiateMsgBytes func([]string, []string, sdk.AccAddress) []byte) cosmwasmpooltypes.CosmWasmExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Set the supply of the denoms, since the contract fails if the denom doesn't exist on chain.
	for _, denom := range denoms {
		err := s.App.BankKeeper.Supply.Set(s.Ctx, denom, osmomath.NewInt(100000000000000))
		s.Require().NoError(err)
	}

	// Upload contract code and get the code id.
	codeId := s.StoreCosmWasmPoolContractCodeCustomProject(contractName, projectName, byteCodePath)

	// Add code id to the whitelist.
	s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, codeId)

	// Generate instantiate message bytes.
	instantiateMsgBz := getInstantiateMsgBytes(denoms, normalizationFactors, owner)

	// Generate msg create pool.
	validCWPoolMsg := model.NewMsgCreateCosmWasmPool(codeId, owner, instantiateMsgBz)

	// Create pool.
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, validCWPoolMsg)
	s.Require().NoError(err)

	// Get and return the pool.
	pool, err := s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	return pool
}

// AddRatioFundsToTransmuterPool adds funds to a transmuter pool based on the provided ratio of denoms.
// It mints the specified amounts of each denom and joins the pool with these funds.
//
// Parameters:
// - owner: The account address that will own the funds.
// - denoms: A list of denominations to be added to the pool.
// - ratioOfDenoms: A list of ratios corresponding to each denom. Must be the same length as denoms.
// - poolId: The ID of the pool to which the funds will be added.
//
// Panics if the length of denoms and ratioOfDenoms are not equal.
func (s *KeeperTestHelper) AddRatioFundsToTransmuterPool(owner sdk.AccAddress, denoms []string, ratioOfDenoms []uint16, poolId uint64) {
	if ratioOfDenoms == nil {
		return
	}

	if len(denoms) != len(ratioOfDenoms) {
		panic("denoms and ratioOfDenoms must be of equal length")
	}

	var poolCoins sdk.Coins
	for i, denom := range denoms {
		amount := osmomath.NewInt(int64(ratioOfDenoms[i])).Mul(osmomath.NewInt(1000000000))
		if amount.IsZero() {
			continue
		}
		poolCoins = append(poolCoins, sdk.NewCoin(denom, amount))
	}

	// Add funds to the pool
	s.FundAcc(owner, poolCoins)
	s.JoinTransmuterPool(s.TestAccs[0], poolId, poolCoins)
}

// GetDefaultTransmuterInstantiateMsgBytes returns the default instantiate message for the transmuter contract
// with DefaultTransmuterDenomA and DefaultTransmuterDenomB as the pool asset denoms.
func (s *KeeperTestHelper) GetDefaultTransmuterInstantiateMsgBytes() []byte {
	return s.GetTransmuterInstantiateMsgBytes([]string{DefaultTransmuterDenomA, DefaultTransmuterDenomB}, nil, sdk.AccAddress{})
}

// GetTransmuterInstantiateMsgBytes returns the instantiate message for the transmuter contract with the
// given pool asset denoms.
func (s *KeeperTestHelper) GetTransmuterInstantiateMsgBytes(poolAssetDenoms []string, normalizationFactors []string, owner sdk.AccAddress) []byte {
	instantiateMsg := msg.InstantiateMsg{
		PoolAssetDenoms: poolAssetDenoms,
	}

	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)

	return instantiateMsgBz
}

// GetTransmuterInstantiateMsgBytesV3 returns the instantiate message for the transmuter contract with the
// given pool asset denoms and their respective normalization factors.
func (s *KeeperTestHelper) GetTransmuterInstantiateMsgBytesV3(poolAssetDenoms []string, normalizationFactors []string, owner sdk.AccAddress) []byte {
	var assetConfigs []v3.AssetConfig
	for i, denom := range poolAssetDenoms {
		assetConfigs = append(assetConfigs, v3.AssetConfig{
			Denom:               denom,
			NormalizationFactor: normalizationFactors[i],
		})
	}

	instantiateMsg := v3.InstantiateMsg{
		PoolAssetConfigs:                assetConfigs,
		AlloyedAssetSubdenom:            "testdenom",
		AlloyedAssetNormalizationFactor: "1",
		Admin:                           owner.String(),
		Moderator:                       owner.String(),
	}

	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	fmt.Println(string(instantiateMsgBz))
	s.Require().NoError(err)

	return instantiateMsgBz
}

// StoreCosmWasmPoolContractCode stores the cosmwasm pool contract code in the wasm keeper and returns the code id.
// contractName is the name of the contract file in the x/cosmwasmpool/bytecode directory without the .wasm extension.
func (s *KeeperTestHelper) StoreCosmWasmPoolContractCode(contractName string) uint64 {
	return s.StoreCosmWasmPoolContractCodeCustomProject(contractName, osmosisRepository, osmosisRepoTransmuterPath)
}

// StoreCosmWasmPoolContractCodeCustomProject stores the cosmwasm pool contract code in the wasm keeper and returns the code id.
// contractName is the name of the contract file in the x/cosmwasmpool/bytecode directory without the .wasm extension.
// It has the flexibility of being used from outside the Osmosis repository by providing custom project name and bytecode path.
func (s *KeeperTestHelper) StoreCosmWasmPoolContractCodeCustomProject(contractName, projectName, byteCodePath string) uint64 {
	cosmwasmpoolModuleAddr := s.App.AccountKeeper.GetModuleAddress(cosmwasmpooltypes.ModuleName)
	s.Require().NotNil(cosmwasmpoolModuleAddr)

	// Allow the cosmwasm pool module to upload code.
	params := s.App.WasmKeeper.GetParams(s.Ctx)
	err := s.App.WasmKeeper.SetParams(s.Ctx, wasmtypes.Params{
		CodeUploadAccess: wasmtypes.AccessConfig{
			Permission: wasmtypes.AccessTypeAnyOfAddresses,
			Addresses:  []string{cosmwasmpoolModuleAddr.String()},
		},
		InstantiateDefaultPermission: params.InstantiateDefaultPermission,
	})
	s.Require().NoError(err)

	code := s.GetContractCodeCustomProject(contractName, projectName, byteCodePath)

	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeAnyOfAddresses, Addresses: []string{cosmwasmpoolModuleAddr.String()}}
	codeID, _, err := s.App.ContractKeeper.Create(s.Ctx, cosmwasmpoolModuleAddr, code, &instantiateConfig)
	s.Require().NoError(err)

	return codeID
}

func (s *KeeperTestHelper) GetContractCode(contractName string) []byte {
	return s.GetContractCodeCustomProject(contractName, "osmosis", "x/cosmwasmpool/bytecode")
}

// GetContractCode returns the contract code for the given contract name.
// Assumes that the contract code is stored under x/cosmwasmpool/bytecode.
func (s *KeeperTestHelper) GetContractCodeCustomProject(contractName string, projectName string, path string) []byte {
	workingDir, err := os.Getwd()
	s.Require().NoError(err)

	projectRootPath := fmt.Sprintf("/%s/", projectName)
	projectRootIndex := strings.LastIndex(workingDir, projectRootPath) + len(projectRootPath)
	workingDir = workingDir[:projectRootIndex]
	code, err := os.ReadFile(workingDir + path + "/" + contractName + ".wasm")
	s.Require().NoError(err)

	return code
}

// JoinTransmuterPool joins the given pool with the given coins from the given address.
func (s *KeeperTestHelper) JoinTransmuterPool(lpAddress sdk.AccAddress, poolId uint64, coins sdk.Coins) {
	pool, err := s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)
	// add liquidity by joining the pool
	request := transmuter.JoinPoolExecuteMsgRequest{}
	cosmwasm.MustExecute[transmuter.JoinPoolExecuteMsgRequest, msg.EmptyStruct](s.Ctx, s.App.ContractKeeper, pool.GetContractAddress(), lpAddress, coins, request)
}
