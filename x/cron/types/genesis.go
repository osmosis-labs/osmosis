package types

import "fmt"

func NewGenesisState(cronJobs []CronJob, params Params) *GenesisState {
	return &GenesisState{
		CronJobs: cronJobs,
		Params:   params,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		[]CronJob{},
		DefaultParams(),
	)
}

func (genState *GenesisState) Validate() error {
	if err := genState.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	// validates all the cron jobs
	for i, cron := range genState.CronJobs {
		if err := cron.Validate(); err != nil {
			return fmt.Errorf("invalid cron %d: %w", i, err)
		}
	}
	return nil
}
