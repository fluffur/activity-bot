package logger

// Truncate returns s if its rune length is at most maxRunes; otherwise a prefix and a marker.
func Truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "…(truncated)"
}
