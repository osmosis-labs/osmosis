package simtypes

type Frequency string

const (
	Rare       = "Rare"
	Infrequent = "Infrequent"
	Common     = "Common"
	Frequent   = "Frequent"
)

func MapFrequencyFromInt(intFrequency int) string {
	switch {
	case intFrequency < 10:
		return "Rare"
	case intFrequency > 10 && intFrequency < 20:
		return "Infrequent"
	case intFrequency > 20 && intFrequency < 50:
		return "Common"
	case intFrequency > 50:
		return "Frequent"
	default:
		return "Common"
	}
}

func MapFrequencyFromString(strFrequency Frequency) int {
	switch {
	case strFrequency == "Rare":
		return 5
	case strFrequency == "Infrequent":
		return 15
	case strFrequency == "Common":
		return 35
	case strFrequency == "Frequent":
		return 65
	default:
		return 35
	}
}
