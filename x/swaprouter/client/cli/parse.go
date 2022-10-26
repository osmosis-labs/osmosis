package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type createPoolInputs struct {
	Weights                  string                         `json:"weights"`
	InitialDeposit           string                         `json:"initial-deposit"`
	SwapFee                  string                         `json:"swap-fee"`
	ExitFee                  string                         `json:"exit-fee"`
	FutureGovernor           string                         `json:"future-governor"`
	SmoothWeightChangeParams smoothWeightChangeParamsInputs `json:"lbp-params"`
}

type smoothWeightChangeParamsInputs struct {
	StartTime         string `json:"start-time"`
	Duration          string `json:"duration"`
	TargetPoolWeights string `json:"target-pool-weights"`
}

type XCreatePoolInputs createPoolInputs

type XCreatePoolInputsExceptions struct {
	XCreatePoolInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createPoolInputs) UnmarshalJSON(data []byte) error {
	var createPoolE XCreatePoolInputsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createPoolE); err != nil {
		return err
	}

	*release = createPoolInputs(createPoolE.XCreatePoolInputs)
	return nil
}

func parseCreatePoolFlags(fs *pflag.FlagSet) (*createPoolInputs, error) {
	pool := &createPoolInputs{}
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
