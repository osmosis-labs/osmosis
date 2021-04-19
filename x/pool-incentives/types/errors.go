package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoPotIdExist                 = sdkerrors.Register(ModuleName, 1, "no pot id exist")
	ErrDistrRecordNotPositiveWeight = sdkerrors.Register(ModuleName, 2, "weight in record should be positive")
	ErrDistrRecordInvalidIndex      = sdkerrors.Register(ModuleName, 3, "invalid index")
	ErrDistrRecordMismatchedPotId   = sdkerrors.Register(ModuleName, 4, "pot id mismatched")

	ErrEmptyProposalRecords = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalIndexes = sdkerrors.Register(ModuleName, 11, "indexes are empty")
)
