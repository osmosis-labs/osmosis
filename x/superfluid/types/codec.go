package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSuperfluidDelegate{}, "symphony/superfluid-delegate", nil)
	cdc.RegisterConcrete(&MsgSuperfluidUndelegate{}, "symphony/superfluid-undelegate", nil)
	cdc.RegisterConcrete(&MsgLockAndSuperfluidDelegate{}, "symphony/lock-and-superfluid-delegate", nil)
	cdc.RegisterConcrete(&MsgSuperfluidUnbondLock{}, "symphony/superfluid-unbond-lock", nil)
	cdc.RegisterConcrete(&MsgSuperfluidUndelegateAndUnbondLock{}, "symphony/sf-undelegate-and-unbond-lock", nil)
	cdc.RegisterConcrete(&SetSuperfluidAssetsProposal{}, "symphony/set-superfluid-assets-proposal", nil)
	cdc.RegisterConcrete(&UpdateUnpoolWhiteListProposal{}, "symphony/update-unpool-whitelist", nil)
	cdc.RegisterConcrete(&RemoveSuperfluidAssetsProposal{}, "symphony/del-superfluid-assets-proposal", nil)
	cdc.RegisterConcrete(&MsgUnPoolWhitelistedPool{}, "symphony/unpool-whitelisted-pool", nil)
	cdc.RegisterConcrete(&MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition{}, "symphony/unlock-and-migrate", nil)
	cdc.RegisterConcrete(&MsgCreateFullRangePositionAndSuperfluidDelegate{}, "symphony/full-range-and-sf-delegate", nil)
	cdc.RegisterConcrete(&MsgAddToConcentratedLiquiditySuperfluidPosition{}, "symphony/add-to-cl-superfluid-position", nil)
	cdc.RegisterConcrete(&MsgUnbondConvertAndStake{}, "symphony/unbond-convert-and-stake", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSuperfluidDelegate{},
		&MsgSuperfluidUndelegate{},
		&MsgLockAndSuperfluidDelegate{},
		&MsgSuperfluidUnbondLock{},
		&MsgSuperfluidUndelegateAndUnbondLock{},
		&MsgUnPoolWhitelistedPool{},
		&MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition{},
		&MsgCreateFullRangePositionAndSuperfluidDelegate{},
		&MsgAddToConcentratedLiquiditySuperfluidPosition{},
		&MsgUnbondConvertAndStake{},
	)

	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&SetSuperfluidAssetsProposal{},
		&RemoveSuperfluidAssetsProposal{},
		&UpdateUnpoolWhiteListProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterCodec(authzcodec.Amino)
	amino.Seal()
}
