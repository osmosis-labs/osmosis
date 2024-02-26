package authenticator

type UsedAuthenticators struct {
	usedAuthenticators []uint64
}

func NewUsedAuthenticators() *UsedAuthenticators {
	return &UsedAuthenticators{
		usedAuthenticators: []uint64{},
	}
}

func (as *UsedAuthenticators) ResetUsedAuthenticators() {
	as.usedAuthenticators = []uint64{}
}

func (as *UsedAuthenticators) GetUsedAuthenticators() []uint64 {
	return as.usedAuthenticators
}

func (as *UsedAuthenticators) AddUsedAuthenticator(authenticatorId uint64) {
	as.usedAuthenticators = append(as.usedAuthenticators, authenticatorId)
}
