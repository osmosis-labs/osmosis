package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

type XCreatePoolInputs createPoolInputs

type XCreatePoolInputsExceptions struct {
	XCreatePoolInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected
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

	contents, err := ioutil.ReadFile(poolFile)
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
