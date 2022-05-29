package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

type CreateSaleInputs createSaleInputs

// UnmarshalJSON should error if there are fields unexpected.
func (inputs *createSaleInputs) UnmarshalJSON(data []byte) error {
	var saleInputs CreateSaleInputs
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&saleInputs); err != nil {
		return err
	}

	*inputs = createSaleInputs(saleInputs)
	return nil
}

func parseCreateSaleFlags(fs *pflag.FlagSet) (*createSaleInputs, error) {
	sale := &createSaleInputs{}
	saleFile, _ := fs.GetString(FlagSaleFile)

	if saleFile == "" {
		return nil, fmt.Errorf("must pass in a sale json using the --%s flag", FlagSaleFile)
	}

	contents, err := ioutil.ReadFile(saleFile)
	if err != nil {
		return nil, err
	}

	// make exception if unknown field exists
	err = sale.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return sale, nil
}
