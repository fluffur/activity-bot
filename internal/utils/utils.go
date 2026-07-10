package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func UcFirst(s string) string {
	if s == "" {
		return ""
	}

	r, size := utf8.DecodeRuneInString(s)

	rUpper := unicode.ToUpper(r)

	var b strings.Builder

	b.Grow(len(s))
	b.WriteRune(rUpper)
	b.WriteString(s[size:])

	return b.String()
}
