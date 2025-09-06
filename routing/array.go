package routing

import (
	"math/rand"
)

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

// Get a random element of s.
// If len(s) is 0, returns nil, false.
func randomElement[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}

	i := rand.Intn(len(s))
	return s[i], true
}
