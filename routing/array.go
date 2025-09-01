package routing

// Remove the element of s at index i.
func removeAt[T any](s []T, i int) []T {
	if i < 0 || i >= len(s) {
		return s
	}
	copy(s[i:], s[i+1:])
	var zero T
	s[len(s)-1] = zero
	return s[:len(s)-1]
}
