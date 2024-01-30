package utils

import "cmp"

func Remove[T cmp.Ordered](s []T, e T) []T {
	for i, el := range s {
		if e == el {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
