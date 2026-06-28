package summon

import "activity-bot/internal/chat"

func mentionSeparator(mt chat.MentionTypes) string {
	if mt.Has(chat.MentionEmoji) && !mt.Has(chat.MentionRole) && !mt.Has(chat.MentionName) {
		return " "
	}
	return ", "
}

func utf16ToRuneIndex(s string, utf16Pos int) int {
	count := 0

	for i, r := range s {
		if count >= utf16Pos {
			return i
		}

		if r > 0xFFFF {
			count += 2
		} else {
			count++
		}
	}

	return len(s)
}

func chunk[T any](items []T, size int) [][]T {
	if size <= 0 {
		return [][]T{items}
	}

	var chunks [][]T

	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}

		chunks = append(chunks, items[i:end])
	}

	return chunks
}
