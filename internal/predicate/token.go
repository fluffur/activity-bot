package predicate

import "unicode/utf16"

func getFreeTokens(s string, used []Offset) []token {
	var tokens []token
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\n' && s[j] != '\r' {
			j++
		}

		overlapping := false
		for _, o := range used {
			if i < o.End && j > o.Start {
				overlapping = true
				break
			}
		}

		if !overlapping {
			tokens = append(tokens, token{text: s[i:j], start: i, end: j})
		}
		i = j
	}
	return tokens
}

func entityTextUTF16(text string, offset16, length16 int) string {
	u16 := utf16.Encode([]rune(text))
	if offset16 < 0 || offset16 >= len(u16) {
		return ""
	}
	end16 := offset16 + length16
	if end16 > len(u16) {
		end16 = len(u16)
	}
	return string(utf16.Decode(u16[offset16:end16]))
}
