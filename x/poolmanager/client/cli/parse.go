package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// TODO: move these to exported types within an internal package
type XCreatePoolInputs createBalancerPoolInputs

type XCreatePoolInputsExceptions struct {
	XCreatePoolInputs
	Other *string // Other won't raise an error
}

type XCreateStableswapPoolInputs createStableswapPoolInputs

type XCreateStableswapPoolInputsExceptions struct {
	XCreateStableswapPoolInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createBalancerPoolInputs) UnmarshalJSON(data []byte) error {
	var createPoolE XCreatePoolInputsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createPoolE); err != nil {
		return err
	}

	*release = createBalancerPoolInputs(createPoolE.XCreatePoolInputs)
	return nil
}

func parseCreateBalancerPoolFlags(fs *pflag.FlagSet) (*createBalancerPoolInputs, error) {
	pool := &createBalancerPoolInputs{}
	poolFile, _ := fs.GetString(FlagPoolFile)

	if poolFile == "" {
		return nil, fmt.Errorf("must pass in a pool json using the --%s flag", FlagPoolFile)
	}

	contents, err := os.ReadFile(poolFile)
	if err != nil {
		return nil, err
	}

	// make exception if unknown field exists
	err = pool.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createStableswapPoolInputs) UnmarshalJSON(data []byte) error {
	var createPoolE XCreateStableswapPoolInputsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createPoolE); err != nil {
		return err
	}

	*release = createStableswapPoolInputs(createPoolE.XCreateStableswapPoolInputs)
	return nil
}

func parseCreateStableswapPoolFlags(fs *pflag.FlagSet) (*createStableswapPoolInputs, error) {
	pool := &createStableswapPoolInputs{}
	poolFile, _ := fs.GetString(FlagPoolFile)

	if poolFile == "" {
		return nil, fmt.Errorf("must pass in a pool json using the --%s flag", FlagPoolFile)
	}

	contents, err := os.ReadFile(poolFile)
	if err != nil {
		return nil, err
	}

	// make exception if unknown field exists
	err = pool.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
