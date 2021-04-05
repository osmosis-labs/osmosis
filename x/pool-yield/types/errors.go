package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoFarmIdExist                = sdkerrors.Register(ModuleName, 1, "no farm id exist")
	ErrDistrRecordNotPositiveWeight = sdkerrors.Register(ModuleName, 2, "weight in record should be positive")

	ErrEmptyProposalRecords = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalIndexes = sdkerrors.Register(ModuleName, 10, "indexes are empty")
)
