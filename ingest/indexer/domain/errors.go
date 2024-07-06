package domain

import "errors"

var ErrDidNotIngestAllData = errors.New("cold start manager has not yet ingested initial data")
