package domain

import "errors"

var ErrColdStartManagerDidNotIngest = errors.New("cold start manager has not yet ingested initial data")
