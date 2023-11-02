package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (q *RegisteredQuery) GetOwnerAddress() (creator sdk.AccAddress, err error) {
	creator, err = sdk.AccAddressFromBech32(q.Owner)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode owner address: %s", q.Owner)
	}

	return creator, nil
}

// ValidateRemoval checks whether the caller is authorized to remove the query in current
// circumstances. Valid cases are:
// 1. owner removes query at any time;
// 2. anyone removes query if there's been q.SubmitTimeout blocks since last result submission
// height and query registration height.
func (q *RegisteredQuery) ValidateRemoval(ctx sdk.Context, caller string) error {
	if q.GetOwner() == caller {
		return nil // query owner is authorized to remove their queries at any time
	}

	registrationTimeoutBlock := q.RegisteredAtHeight + q.SubmitTimeout
	submitTimeoutBlock := q.LastSubmittedResultLocalHeight + q.SubmitTimeout
	currentBlock := uint64(ctx.BlockHeader().Height)
	if currentBlock <= registrationTimeoutBlock || currentBlock <= submitTimeoutBlock {
		return fmt.Errorf("only owner can remove a query within its service period")
	}
	return nil
}
