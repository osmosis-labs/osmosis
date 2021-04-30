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
		pool.Weights, _ = fs.GetString(FlagWeights)
		pool.InitialDeposit, _ = fs.GetString(FlagInitialDeposit)
		pool.SwapFee, _ = fs.GetString(FlagSwapFee)
		pool.ExitFee, _ = fs.GetString(FlagExitFee)
		pool.FutureGovernor, _ = fs.GetString(FlagFutureGovernor)
		return pool, nil
	}

	for _, flag := range CreatePoolFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --%s, which is a noop", flag, FlagPoolFile)
		}
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
