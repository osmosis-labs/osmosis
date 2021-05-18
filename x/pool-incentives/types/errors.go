package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoPotIdExist                 = sdkerrors.Register(ModuleName, 1, "no pot id exist")
	ErrDistrRecordNotPositiveWeight = sdkerrors.Register(ModuleName, 2, "weight in record should be positive")
	ErrDistrRecordNotRegisteredPot  = sdkerrors.Register(ModuleName, 3, "pot was not registered")
	ErrDistrRecordRegisteredPot     = sdkerrors.Register(ModuleName, 4, "pot was already registered")

	ErrEmptyProposalRecords = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalPotIds  = sdkerrors.Register(ModuleName, 11, "pot ids are empty")
)
