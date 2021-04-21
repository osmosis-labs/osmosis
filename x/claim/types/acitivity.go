package types

func ActionToNames(actions []Action) []string {
	names := []string{}
	for _, action := range actions {
		names = append(names, Action_name[action])
	}
	return names
}