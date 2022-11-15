package types

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func Intersection(a, b *[]string) []string {
	var result []string
	for _, v := range *a {
		if Contains(*b, v) {
			result = append(result, v)
		}
	}
	return result
}
