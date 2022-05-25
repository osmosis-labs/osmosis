package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

type CreateLBPInputs createLBPInputs

type XCreateLBPInputsExceptions struct {
	CreateLBPInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (inputs *createLBPInputs) UnmarshalJSON(data []byte) error {
	var createLBPE XCreateLBPInputsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createLBPE); err != nil {
		return err
	}

	*inputs = createLBPInputs(createLBPE.CreateLBPInputs)
	return nil
}

func parseCreateLBPFlags(fs *pflag.FlagSet) (*createLBPInputs, error) {
	lbp := &createLBPInputs{}
	lbpFile, _ := fs.GetString(FlagLBPFile)

	if lbpFile == "" {
		return nil, fmt.Errorf("must pass in a lbp json using the --%s flag", FlagLBPFile)
	}

	contents, err := ioutil.ReadFile(lbpFile)
	if err != nil {
		return nil, err
	}

	// make exception if unknown field exists
	err = lbp.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return lbp, nil
}
