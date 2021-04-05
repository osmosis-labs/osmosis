package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoFarmIdExist                = sdkerrors.Register(ModuleName, 1, "no farm id exist")
	ErrDistrRecordNotPositiveWeight = sdkerrors.Register(ModuleName, 2, "weight in record should be positive")
	ErrDistrRecordInvalidIndex      = sdkerrors.Register(ModuleName, 3, "invalid index")

	ErrEmptyProposalRecords = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalIndexes = sdkerrors.Register(ModuleName, 11, "indexes are empty")
)
