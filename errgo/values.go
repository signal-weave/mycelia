package errgo

import "fmt"

// ValueOrPanic panics if the error of a function that returns a tuple (val, err)
// returned a non-nil error, otherwise will return the value.
func ValueOrPanic[T any](value T, err error) T {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return value
}

// PanicIfError panics if err is not nil.
// For use with functions that only return an error or nil.
func PanicIfError(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
