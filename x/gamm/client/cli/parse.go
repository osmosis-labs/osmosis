package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
)

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

	err = json.Unmarshal(contents, pool)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
