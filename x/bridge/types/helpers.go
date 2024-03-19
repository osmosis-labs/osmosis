package types

// Map returns a slice that contains the result of applying function f
// to every element of the slice s.
// TODO: Placed here temporarily. Delete after releasing the new osmoutils version.
func Map[E, V any](s []E, f func(E) V) []V {
	res := make([]V, 0, len(s))
	for _, v := range s {
		res = append(res, f(v))
	}
	return res
}
