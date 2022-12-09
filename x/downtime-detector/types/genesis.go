package types

import "time"

func DefaultGenesis() *GenesisState {
	genDowntimes := []GenesisDowntimeEntry{}
	for _, downtime := range DowntimeToDuration.Keys() {
		genDowntimes = append(genDowntimes, GenesisDowntimeEntry{
			Duration:     downtime,
			LastDowntime: time.Unix(0, 0),
		})
	}
	return &GenesisState{
		Downtimes:     genDowntimes,
		LastBlockTime: time.Unix(0, 0),
	}
}

func (g *GenesisState) Validate() error {
	return nil
}
