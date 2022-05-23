package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

type XCreateLBPInputs createLBPInputs

type XCreateLBPInputsExceptions struct {
	XCreateLBPInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createLBPInputs) UnmarshalJSON(data []byte) error {
	var createLBPE XCreateLBPInputsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createLBPE); err != nil {
		return err
	}

	*release = createLBPInputs(createLBPE.XCreateLBPInputs)
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
