package types

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

func validateStrLen(s, field string, min, max int, errmsgs []string) []string {
	if len(s) < min || len(s) > max {
		errmsgs = append(errmsgs, fmt.Sprintf("%q length must be between <%d, %d>", s, min, max))
	}
	return errmsgs
}
