package keeper

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
	"golang.org/x/exp/slices"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		paramstore paramtypes.Subspace
		conOps     types.ContractOpsKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	conOps types.ContractOpsKeeper,

) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramstore: ps,
		conOps:     conOps,
	}
}

//nolint:staticcheck
func (k Keeper) SudoContractCall(ctx sdk.Context, contractAddress string, p []byte) error {
	contractAddr, err := sdk.AccAddressFromBech32(contractAddress)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, contractAddress)
	}
	data, err := k.conOps.Sudo(ctx, contractAddr, p)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeContractSudoMsg,
		sdk.NewAttribute(types.AttributeKeyResultDataHex, hex.EncodeToString(data)),
	))
	return nil
}

func (k Keeper) CheckSecurityAddress(ctx sdk.Context, from string) bool {
	params := k.GetParams(ctx)
	return slices.Contains(params.SecurityAddress, from)
}

func (k Keeper) Store(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}
