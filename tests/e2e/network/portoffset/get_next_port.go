package portoffset

var next int

func init() {
	next = 10
}

func GetNext() int {
	nextToReturn := next
	next += 10
	return nextToReturn
}
