package cosmwasmpool

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

func NewCosmWasmPoolProposalHandler(k Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.UploadCosmWasmPoolCodeAndWhiteListProposal:
			_, err := k.uploadCodeIdAndWhitelist(ctx, c.WASMByteCode)
			return err
		case *types.MigratePoolContractsProposal:
			return k.migrateCosmwasmPools(ctx, c.PoolIds, c.NewCodeId, c.WASMByteCode, c.MigrateMsg)
		default:
			return fmt.Errorf("unrecognized concentrated liquidity proposal content type: %T", c)
		}
	}
}

// uploadCodeIdAndWhitelist uploads the given wasm bytecode to the wasmvm. Whitelists the resulting code id
// Emits an event with the code id and checksum.
// Returns error if byte code is empty or if fails to upload the code.
func (k Keeper) uploadCodeIdAndWhitelist(ctx sdk.Context, byteCode []byte) (uint64, error) {
	if len(byteCode) == 0 {
		return 0, fmt.Errorf("empty wasm bytecode")
	}

	cosmwasmPoolModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Only allow the x/cosmwasmpool module to instantiate this contract.
	instantiatePermissions := wasmtypes.AccessConfig{
		Permission: wasmtypes.AccessTypeAnyOfAddresses,
		Addresses:  []string{cosmwasmPoolModuleAddress.String()},
	}

	// Upload the code to the wasmvm.
	codeID, checksum, err := k.contractKeeper.Create(ctx, cosmwasmPoolModuleAddress, byteCode, &instantiatePermissions)
	if err != nil {
		return 0, err
	}

	// Add the code id to the whitelist.
	k.WhitelistCodeId(ctx, codeID)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.TypeEvtUploadedCosmwasmPoolCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)),
		sdk.NewAttribute(types.AttributeKeyChecksum, string(checksum)),
	))

	return codeID, nil
}

// migrateComswasmPools migrates all given cw pool contracts specified by their IDs.
// It has two options to perform the migration.
//
// 1. If the codeID is non-zero, it will migrate the pool contracts to a given codeID assuming that it has already
// been uploaded. uploadByteCode must be empty in such a case. Fails if codeID does not exist.
// Fails if uploadByteCode is not empty.
//
// 2. If the codeID is zero, it will upload the given uploadByteCode and use the new resulting code id to migrate
// the pool to. Errors if uploadByteCode is empty or invalid.
//
// In both cases, if one of the pools specified by the given poolID does not exist, the proposal fails.
//
// The reason for having poolIDs be a slice of ids is to account for the potential need for emergency migration
// of all old code ids associated with particular pools to new code ids, or simply having the flexibility of
// migrating multiple older pool contracts to a new one at once when there is a release.
//
// poolD count to be submitted at once is gated by a governance paramets (20 at launch).
// The proposal fails if more. Note that 20 was chosen arbitrarily to have a constant bound on the number of pools migrated
// at once. This size will be configured by a module parameter so it can be changed by a constant.
func (k Keeper) migrateCosmwasmPools(ctx sdk.Context, poolIds []uint64, newCodeId uint64, uploadByteCode []byte, migrateMsg []byte) (err error) {
	cosmwasmPoolModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)

	if err := types.ValidateMigrationProposalConfiguration(poolIds, newCodeId, uploadByteCode); err != nil {
		return err
	}

	// Validate that the given pool ids are below the pool count limit.
	requestedPoolMigrationCount := uint64(len(poolIds))
	params := k.GetParams(ctx)
	poolMigrationLimit := params.PoolMigrationLimit
	if requestedPoolMigrationCount > poolMigrationLimit {
		return fmt.Errorf("pool migration count (%d) exceeds limit (%d)", requestedPoolMigrationCount, poolMigrationLimit)
	}

	// Iterate requested pool ids to make sure that pool with such id exists.
	poolCount := k.poolmanagerKeeper.GetNextPoolId(ctx) - 1
	for _, poolId := range poolIds {
		if poolId > poolCount {
			return fmt.Errorf("pool id (%d) does not exist", poolId)
		}
	}

	// Upload code id and whitelist it if uploadByteCode is given.
	// Set newCodeId to the resulting code id.
	if len(uploadByteCode) > 0 {
		newCodeId, err = k.uploadCodeIdAndWhitelist(ctx, uploadByteCode)
		if err != nil {
			return err
		}
	}

	// Iterate over pool ids and attempt to migrate each pool's contract.
	for _, poolId := range poolIds {
		cwPool, err := k.GetPoolById(ctx, poolId)
		if err != nil {
			return err
		}

		_, err = k.contractKeeper.Migrate(ctx, sdk.MustAccAddressFromBech32(cwPool.GetContractAddress()), cosmwasmPoolModuleAddress, newCodeId, migrateMsg)
		if err != nil {
			return err
		}

		// Update code ID to the updated one in state
		cwPool.SetCodeId(newCodeId)
		k.SetPool(ctx, cwPool)
	}

	// Whitelist new code id. No-op if already whitelisted.
	k.WhitelistCodeId(ctx, newCodeId)

	// Emit event.
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.TypeEvtMigratedCosmwasmPoolCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(newCodeId, 10)),
		sdk.NewAttribute(types.AttributeKeyPoolIDsMigrated, fmt.Sprintf("%v", poolIds))))

	return nil
}
