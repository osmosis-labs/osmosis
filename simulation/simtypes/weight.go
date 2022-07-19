package simtypes

type Weight int64

const (
	Undefined  Weight = 0
	Rare              = 1
	Infrequent        = 5
	Normal            = 10
	Frequent          = 20
)
