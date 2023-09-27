package accum

import "github.com/osmosis-labs/osmosis/osmomath"

var one = osmomath.OneDec()

// validate returns nil if Options are valid.
// Error otherwise. Note, that, currently,
// options do not contain any fields. We
// create them to be able to extend in the
// future with auto-compounding logic.
// As a result, this always returns nil.
func (o *Options) validate() error {
	if o == nil {
		return nil
	}
	return nil
}
