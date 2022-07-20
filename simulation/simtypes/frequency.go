package simtypes

type Frequency string

const (
	Rare       = "Rare"
	Infrequent = "Infrequent"
	Common     = "Common"
	Frequent   = "Frequent"
)

func mapFrequencyFromInt(intFrequency int) string {
	switch {
	case intFrequency < 10:
		return Rare
	case intFrequency > 10 && intFrequency < 20:
		return Infrequent
	case intFrequency > 20 && intFrequency < 50:
		return Common
	case intFrequency > 50:
		return Frequent
	default:
		return Common
	}
}

func mapFrequencyFromString(strFrequency Frequency) int {
	switch strFrequency {
	case Rare:
		return 5
	case Infrequent:
		return 15
	case Common:
		return 35
	case Frequent:
		return 65
	default:
		return 35
	}
}
