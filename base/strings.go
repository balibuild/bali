package base

import (
	"errors"
	"strings"
)

// string base

// StrSplitSkipEmpty skip empty string suggestcap is suggest cap
func StrSplitSkipEmpty(s string, sep byte, suggestcap int) []string {
	sv := make([]string, 0, suggestcap)
	var first, i int
	for ; i < len(s); i++ {
		if s[i] != sep {
			continue
		}
		if first != i {
			sv = append(sv, s[first:i])
		}
		first = i + 1
	}
	if first < len(s) {
		sv = append(sv, s[first:])
	}
	return sv
}

// StrCat cat strings:
// You should know that StrCat gradually builds advantages
// only when the number of parameters is> 2.
func StrCat(sv ...string) string {
	var sb strings.Builder
	var size int
	for _, s := range sv {
		size += len(s)
	}
	sb.Grow(size)
	for _, s := range sv {
		_, _ = sb.WriteString(s)
	}
	return sb.String()
}

// ErrorCat todo
func ErrorCat(sv ...string) error {
	return errors.New(StrCat(sv...))
}
