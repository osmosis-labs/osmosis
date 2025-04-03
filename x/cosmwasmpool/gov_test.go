package cosmwasmpool_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

type CWPoolGovSuite struct {
	apptesting.KeeperTestHelper
}

func TestCWPoolGovSuite(t *testing.T) {
	suite.Run(t, new(CWPoolGovSuite))
}

// TestUploadCodeIdAndWhitelist tests that the core proposal logic for uploading
// contract code and whitelisting it works as expected.
// It does not test specific errors returned as it is non-trivial to set up
// due to dependency on the wasm keeper and contract keeper that do not export
// errors correctly.
// Test vectors considered:
// 1. Cosmwasm pool module does not have upload access - error.
// 2. New byte code uploaded - success and whitelisted.
// 3. Same byte code uploaded - success and whitelisted but new code id created.
// 4. Empty byte code uploaded - error.
// 5. Invalid byte code uploaded - error.
// 6. For success cases, tests that relevant event is emitted.
func (s *CWPoolGovSuite) TestUploadCodeIdAndWhitelist() {

	// Note that setup is done once and the state is shared between the test cases.
	s.Setup()

	// Get valid transmuter code.
	validTransmuterCode := s.GetContractCode(apptesting.TransmuterContractName)

	tests := []struct {
		name                                     string
		byteCode                                 []byte
		expectedCodeId                           uint64
		shouldWhitelistCWPoolModuleAccountUpload bool
		expectedErr                              bool
	}{
		{
			name:           "error: cw pool module account does not have upload access",
			byteCode:       validTransmuterCode,
			expectedCodeId: validCodeId,
			expectedErr:    true,
		},
		{
			name:     "happy path",
			byteCode: validTransmuterCode,
			// Note that subsequent test cases inherit the whitelist being configured
			// as all test cases share initial setup and state.
			shouldWhitelistCWPoolModuleAccountUpload: true,
			expectedCodeId:                           validCodeId,
		},
		{
			name:           "happy path, same contract code, different code id",
			byteCode:       validTransmuterCode,
			expectedCodeId: validCodeId + 1,
		},
		{
			name:        "error: empty byte code",
			byteCode:    []byte{},
			expectedErr: true,
		},
		{
			name:        "error: invalid byte code",
			byteCode:    []byte{0x00, 0x01, 0x02, 0x03},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// Reset the event manager for each test case.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())

			// Change upload permissions for cw pool module
			if tc.shouldWhitelistCWPoolModuleAccountUpload {
				wasmKeeperParams := s.App.WasmKeeper.GetParams(s.Ctx)
				cwPoolModuleAddress := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				addressesAllowedCodeUpload := wasmKeeperParams.CodeUploadAccess.Addresses
				wasmKeeperParams.CodeUploadAccess.Permission = wasmtypes.AccessTypeAnyOfAddresses
				wasmKeeperParams.CodeUploadAccess.Addresses = append(addressesAllowedCodeUpload, cwPoolModuleAddress.String())
				s.App.WasmKeeper.SetParams(s.Ctx, wasmKeeperParams)
			}

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// System under test.
			codeId, err := cosmwasmPoolKeeper.UploadCodeIdAndWhitelist(s.Ctx, tc.byteCode)

			if tc.expectedErr {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)

			// Check that the code id is whitelisted.
			s.Require().True(cosmwasmPoolKeeper.IsWhitelisted(s.Ctx, codeId))

			// Validate that the code id is the expected one.
			s.Require().Equal(tc.expectedCodeId, codeId)

			// Validate that the event is emitted.
			s.AssertEventEmitted(s.Ctx, types.TypeEvtUploadedCosmwasmPoolCode, 1)
		})
	}
}

// TestUploadCodeIdAndWhitelist tests that core proposal logic for migrating cosmwasm pools
// works as expected. In particular, we test that the two migration strategies work as expected.
// 1. Migrate given pools to a pre-uploaded code id.
// 2. Migrate given pools to a new byte code.
//
// It does not test specific errors returned as it is non-trivial to set up
// due to dependency on the wasm keeper and contract keeper that do not export
// errors correctly.
//
// Test vectors considered:
// 1. Migration to a pre-uploaded code id works as expected and whitelist updated.
// 2. Migration to a new byte code works as expected and whitelist updated.
// 3. Migration fails because the contract to migrate has no migrate entrypoint.
// 3. Migration fails because the code id is not whitelisted.
// 3. Migration fails because one of the given pool ids does not exist
// 4. Migration fails because more than the limit of allowed pools is attempted to migrate.
// 5. Migration fails because pool id list is empty
// 6. For success cases, tests that relevant event is emitted.
func (s *CWPoolGovSuite) TestMigrateCosmwasmPools() {
	// Get valid transmuter code.
	validTransmuterCodeNoMigrateEntrypoint := s.GetContractCode(apptesting.TransmuterContractName)
	// Get valid transmuter code with migration entrypoint.
	validTransmuterMigrateCode := s.GetContractCode(apptesting.TransmuterMigrateContractName)

	const (
		preUploadCodeIdPlaceholder  uint64 = 1000
		zeroCodeId                  uint64 = 0
		defaultPoolCountToPreCreate uint64 = 3

		// We create a code id for each pool. Since we provide code to upload in this test,
		// we expect the code id to be one greater than the default pool count.
		expectedNewCodeId uint64 = defaultPoolCountToPreCreate + 1
	)

	type MigrateMsg struct{}
	migrateMsg := MigrateMsg{}

	emptyMigrateMsg, err := json.Marshal(migrateMsg)
	s.Require().NoError(err)

	var (
		emptyByteCode           []byte = []byte{}
		defaultPoolIdsToMigrate        = []uint64{1, 2, 3}
	)

	tests := []struct {
		name                                     string
		poolCountToPreCreate                     uint64
		poolIdsToMigrate                         []uint64
		newCodeId                                uint64
		byteCode                                 []byte
		migrateMsg                               []byte
		expectedCodeId                           uint64
		shouldWhitelistCWPoolModuleAccountUpload bool
		poolIdLimitOverwrite                     uint64

		expectedErr bool
	}{
		{
			name:                 "happy path with pre-uploaded code id",
			poolCountToPreCreate: defaultPoolCountToPreCreate,
			poolIdsToMigrate:     defaultPoolIdsToMigrate,
			newCodeId:            preUploadCodeIdPlaceholder,
			byteCode:             emptyByteCode,
			migrateMsg:           emptyMigrateMsg,

			expectedCodeId: expectedNewCodeId,
		},
		{
			name:                                     "happy path with code id to upload",
			poolCountToPreCreate:                     defaultPoolCountToPreCreate,
			poolIdsToMigrate:                         defaultPoolIdsToMigrate,
			newCodeId:                                zeroCodeId,
			byteCode:                                 validTransmuterMigrateCode,
			migrateMsg:                               emptyMigrateMsg,
			shouldWhitelistCWPoolModuleAccountUpload: true,

			expectedCodeId: expectedNewCodeId,
		},
		{
			name:                                     "error: contract without migration entrypoint",
			poolCountToPreCreate:                     defaultPoolCountToPreCreate,
			poolIdsToMigrate:                         defaultPoolIdsToMigrate,
			newCodeId:                                zeroCodeId,
			byteCode:                                 validTransmuterCodeNoMigrateEntrypoint,
			migrateMsg:                               emptyMigrateMsg,
			shouldWhitelistCWPoolModuleAccountUpload: true,

			expectedErr: true,
		},
		{
			name:                 "error: cw pool module account does not have upload access",
			poolCountToPreCreate: defaultPoolCountToPreCreate,
			poolIdsToMigrate:     defaultPoolIdsToMigrate,
			newCodeId:            zeroCodeId,
			byteCode:             validTransmuterCodeNoMigrateEntrypoint,
			migrateMsg:           emptyMigrateMsg,

			expectedErr: true,
		},
		{
			name:                 "error: migration fails because one of the given pool ids does not exist",
			poolCountToPreCreate: defaultPoolCountToPreCreate,
			poolIdsToMigrate:     append(defaultPoolIdsToMigrate, 4),
			newCodeId:            validCodeId,
			byteCode:             emptyByteCode,
			migrateMsg:           emptyMigrateMsg,

			expectedErr: true,
		},
		{
			name:                 "error: migration fails because pool limit is exceeded",
			poolCountToPreCreate: defaultPoolCountToPreCreate,
			poolIdsToMigrate:     append(defaultPoolIdsToMigrate, 4),
			newCodeId:            validCodeId,
			byteCode:             emptyByteCode,
			migrateMsg:           emptyMigrateMsg,
			poolIdLimitOverwrite: 2,

			expectedErr: true,
		},
		{
			name:                 "error: migration fails because pool id list is empty",
			poolCountToPreCreate: defaultPoolCountToPreCreate,
			poolIdsToMigrate:     []uint64{},
			newCodeId:            preUploadCodeIdPlaceholder,
			byteCode:             emptyByteCode,
			migrateMsg:           emptyMigrateMsg,

			expectedErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.Setup()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// Reset the event manager for each test case.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())

			// Create pools to migrate.
			for i := uint64(0); i < tc.poolCountToPreCreate; i++ {
				s.PrepareCosmWasmPool()
			}

			// Change upload permissions to desired for cw pool module
			// Note that by default the comswasm pool module account is whitelisted
			// in PrepareCosmWasmPool
			if !tc.shouldWhitelistCWPoolModuleAccountUpload {
				wasmKeeperParams := s.App.WasmKeeper.GetParams(s.Ctx)
				wasmKeeperParams.CodeUploadAccess.Permission = wasmtypes.AccessTypeNobody
				s.App.WasmKeeper.SetParams(s.Ctx, wasmKeeperParams)
			}

			// Overwrite pool id limit if needed.
			if tc.poolIdLimitOverwrite != 0 {
				params := cosmwasmPoolKeeper.GetParams(s.Ctx)
				params.PoolMigrationLimit = tc.poolIdLimitOverwrite
				cosmwasmPoolKeeper.SetParams(s.Ctx, params)
			}

			// If the code id is a placeholder, then upload the transmuter code
			// and set tc.newCodeId to the resulting code id.
			if tc.newCodeId == preUploadCodeIdPlaceholder {
				tc.newCodeId = s.StoreCosmWasmPoolContractCode(apptesting.TransmuterMigrateContractName)
			}

			// System under test.
			err := cosmwasmPoolKeeper.MigrateCosmwasmPools(s.Ctx, tc.poolIdsToMigrate, tc.newCodeId, tc.byteCode, tc.migrateMsg)

			if tc.expectedErr {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)

			// Check that the code id is whitelisted.
			s.Require().True(cosmwasmPoolKeeper.IsWhitelisted(s.Ctx, tc.expectedCodeId))

			s.Require().NotEqual(0, len(tc.poolIdsToMigrate))
			for _, poolID := range tc.poolIdsToMigrate {
				// Check that the pool is migrated.
				pool, err := cosmwasmPoolKeeper.GetPoolById(s.Ctx, poolID)
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedCodeId, pool.GetCodeId())
			}

			// Validate that the event is emitted.
			s.AssertEventEmitted(s.Ctx, types.TypeEvtMigratedCosmwasmPoolCode, 1)
		})
	}
}
