package executortypes

import "github.com/osmosis-labs/osmosis/v16/simulation/simtypes"

func totalFrequency(actions []simtypes.ActionsWithMetadata) int {
	totalFrequency := 0
	for _, action := range actions {
		totalFrequency += mapFrequencyFromString(action.Frequency())
	}

	return totalFrequency
}

func mapFrequencyFromInt(intFrequency int) simtypes.Frequency {
	switch {
	case intFrequency < 10:
		return simtypes.Rare
	case intFrequency > 10 && intFrequency < 20:
		return simtypes.Infrequent
	case intFrequency > 20 && intFrequency < 50:
		return simtypes.Common
	case intFrequency > 50:
		return simtypes.Frequent
	default:
		return simtypes.Common
	}
}

func mapFrequencyFromString(strFrequency simtypes.Frequency) int {
	switch strFrequency {
	case simtypes.Rare:
		return 5
	case simtypes.Infrequent:
		return 15
	case simtypes.Common:
		return 35
	case simtypes.Frequent:
		return 65
	default:
		return 35
	}
}
