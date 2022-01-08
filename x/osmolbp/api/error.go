package api

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/types/errors"
)

func errorStringsToError(errmsgs []string) error {
	if len(errmsgs) != 0 {
		return errors.Wrap(errors.ErrInvalidRequest, fmt.Sprintf("%v", errmsgs))
	}
	return nil
}
