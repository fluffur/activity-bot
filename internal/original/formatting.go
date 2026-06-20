package original

import (
	"html"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/gotd/botapi"
)

var mdMap = map[botapi.MessageEntityType]string{
	botapi.EntityBold:   "*",
	botapi.EntityItalic: "_",
	botapi.EntityCode:   "`",
}

var mdV2Map = map[botapi.MessageEntityType]string{
	botapi.EntityBold:                 "*",
	botapi.EntityItalic:               "_",
	botapi.EntityCode:                 "`",
	botapi.EntityPre:                  "```",
	botapi.EntityUnderline:            "__",
	botapi.EntityStrikethrough:        "~",
	botapi.EntitySpoiler:              "||",
	botapi.EntityBlockquote:           ">",
	botapi.EntityExpandableBlockquote: "**>",
}

var htmlMap = map[botapi.MessageEntityType]string{
	botapi.EntityBold:                 "b",
	botapi.EntityItalic:               "i",
	botapi.EntityCode:                 "code",
	botapi.EntityPre:                  "pre",
	botapi.EntityUnderline:            "u",
	botapi.EntityStrikethrough:        "s",
	botapi.EntitySpoiler:              "span class=\"tg-spoiler\"",
	botapi.EntityBlockquote:           "blockquote",
	botapi.EntityExpandableBlockquote: "blockquote expandable",
}

// OriginalMD gets the original markdown formatting of a message text.
func OriginalMD(m botapi.Message) string {
	return getOrigMsgMD(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalMDV2 gets the original markdownV2 formatting of a message text.
func OriginalMDV2(m botapi.Message) string {
	return getOrigMsgMDV2(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalHTML gets the original HTML formatting of a message text.
func OriginalHTML(m botapi.Message) string {
	return getOrigMsgHTML(utf16.Encode([]rune(m.Text)), m.Entities)
}

// OriginalCaptionMD gets the original markdown formatting of a message caption.
func OriginalCaptionMD(m botapi.Message) string {
	return getOrigMsgMD(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalCaptionMDV2 gets the original markdownV2 formatting of a message caption.
func OriginalCaptionMDV2(m botapi.Message) string {
	return getOrigMsgMDV2(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalCaptionHTML gets the original HTML formatting of a message caption.
func OriginalCaptionHTML(m botapi.Message) string {
	return getOrigMsgHTML(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalTextMD gets the original markdown formatting of a message text or caption.
func OriginalTextMD(m botapi.Message) string {
	if m.Text != "" {
		return getOrigMsgMD(utf16.Encode([]rune(m.Text)), m.Entities)
	}

	return getOrigMsgMD(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalTextMDV2 gets the original markdownV2 formatting of a message text or caption.
func OriginalTextMDV2(m botapi.Message) string {
	if m.Text != "" {
		return getOrigMsgMDV2(utf16.Encode([]rune(m.Text)), m.Entities)
	}

	return getOrigMsgMDV2(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// OriginalTextHTML gets the original HTML formatting of a message text caption.
func OriginalTextHTML(m botapi.Message) string {
	if m.Text != "" {
		return getOrigMsgHTML(utf16.Encode([]rune(m.Text)), m.Entities)
	}

	return getOrigMsgHTML(utf16.Encode([]rune(m.Caption)), m.CaptionEntities)
}

// Does not support nesting. only look at upper entities.
func getOrigMsgMD(utf16Data []uint16, ents []botapi.MessageEntity) string {
	out := strings.Builder{}
	prev := 0

	for _, ent := range getUpperEntities(ents) {
		newPrev := ent.Offset + ent.Length
		prevText := string(utf16.Decode(utf16Data[prev:ent.Offset]))

		text := utf16.Decode(utf16Data[ent.Offset:newPrev])
		pre, cleanCntnt, post := splitEdgeWhitespace(string(text), ent)
		cleanCntntRune := []rune(cleanCntnt)

		switch ent.Type {
		case botapi.EntityBold, botapi.EntityItalic, botapi.EntityCode:
			out.WriteString(prevText + pre + mdMap[ent.Type] + escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
		case botapi.EntityPre:
			if ent.Language == "" {
				out.WriteString(prevText + pre + mdMap[ent.Type] +
					escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
			} else {
				out.WriteString(prevText + pre + mdMap[ent.Type] +
					ent.Language + "\n" + escapeContainedMDV1(cleanCntntRune, []rune(mdMap[ent.Type])) + mdMap[ent.Type] + post)
			}
		case botapi.EntityTextMention:
			out.WriteString(prevText + pre + "[" + escapeContainedMDV1(cleanCntntRune, []rune("[]()")) + "](tg://user?id=" +
				strconv.FormatInt(ent.User.ID, 10) + ")" + post)
		case botapi.EntityTextLink:
			out.WriteString(prevText + pre + "[" + escapeContainedMDV1(cleanCntntRune, []rune("[]()")) + "](" + ent.URL + ")" + post)
		default:
			continue
		}

		prev = newPrev
	}

	out.WriteString(string(utf16.Decode(utf16Data[prev:])))

	return out.String()
}

func getOrigMsgHTML(utf16Data []uint16, ents []botapi.MessageEntity) string {
	if len(ents) == 0 {
		return html.EscapeString(string(utf16.Decode(utf16Data)))
	}

	bd := strings.Builder{}
	prev := 0

	for _, e := range getUpperEntities(ents) {
		data, end := fillNestedHTML(utf16Data, e, prev, getChildEntities(e, ents))
		bd.WriteString(data)

		prev = end
	}

	bd.WriteString(html.EscapeString(string(utf16.Decode(utf16Data[prev:]))))

	return bd.String()
}

func getOrigMsgMDV2(utf16Data []uint16, ents []botapi.MessageEntity) (origMsg string) {
	if len(ents) == 0 {
		return string(utf16.Decode(utf16Data))
	}

	bd := strings.Builder{}
	prev := 0

	for _, e := range getUpperEntities(ents) {
		data, end := fillNestedMarkdownV2(utf16Data, e, prev, getChildEntities(e, ents))
		bd.WriteString(data)

		prev = end
	}

	bd.WriteString(string(utf16.Decode(utf16Data[prev:])))

	return bd.String()
}

func fillNestedHTML(data []uint16, ent botapi.MessageEntity, start int, entities []botapi.MessageEntity) (finalHTML string, entEnd int) {
	entEnd = ent.Offset + ent.Length
	if len(entities) == 0 || entEnd < entities[0].Offset {
		// no nesting; just return straight away and move to next.
		return writeFinalHTML(data, ent, start, html.EscapeString(string(utf16.Decode(data[ent.Offset:entEnd])))), entEnd
	}

	subPrev := ent.Offset
	subEnd := ent.Offset
	bd := strings.Builder{}

	for _, e := range getUpperEntities(entities) {
		if e.Offset < subEnd || e == ent {
			continue
		}

		if e.Offset >= entEnd {
			break
		}

		out, end := fillNestedHTML(data, e, subPrev, getChildEntities(e, entities))
		bd.WriteString(out)

		subPrev = end
	}

	bd.WriteString(html.EscapeString(string(utf16.Decode(data[subPrev:entEnd]))))

	return writeFinalHTML(data, ent, start, bd.String()), entEnd
}

func fillNestedMarkdownV2(
	data []uint16,
	ent botapi.MessageEntity,
	start int,
	entities []botapi.MessageEntity,
) (finalMD string, entEnd int) {
	entEnd = ent.Offset + ent.Length
	if len(entities) == 0 || entEnd < entities[0].Offset {
		// no nesting; just return straight away and move to next.
		return writeFinalMarkdownV2(data, ent, start, string(utf16.Decode(data[ent.Offset:entEnd]))), entEnd
	}

	subPrev := ent.Offset
	subEnd := ent.Offset
	bd := strings.Builder{}

	for _, e := range getUpperEntities(entities) {
		if e.Offset < subEnd || e == ent {
			continue
		}

		if e.Offset >= entEnd {
			break
		}

		out, end := fillNestedMarkdownV2(data, e, subPrev, getChildEntities(e, entities))
		bd.WriteString(out)

		subPrev = end
	}

	bd.WriteString(string(utf16.Decode(data[subPrev:entEnd])))

	return writeFinalMarkdownV2(data, ent, start, bd.String()), entEnd
}

func writeFinalHTML(data []uint16, ent botapi.MessageEntity, start int, cntnt string) string {
	prevText := html.EscapeString(string(utf16.Decode(data[start:ent.Offset])))
	switch ent.Type {
	case botapi.EntityBold, botapi.EntityItalic, botapi.EntityCode, botapi.EntityUnderline, botapi.EntityStrikethrough, botapi.EntitySpoiler:
		return prevText + "<" + htmlMap[ent.Type] + ">" + cntnt + "</" + closeHTMLTag(htmlMap[ent.Type]) + ">"
	case botapi.EntityPre:
		if ent.Language == "" {
			return prevText + "<pre>" + cntnt + "</pre>"
		}

		return prevText + `<pre><code class="` + ent.Language + `">` + cntnt + "</code></pre>"
	case botapi.EntityCustomEmoji:
		return prevText + `<tg-emoji emoji-id="` + ent.CustomEmojiID + `">` + cntnt + "</tg-emoji>"
	case botapi.EntityDateTime:
		if ent.DateTimeFormat != "" {
			return prevText + `<tg-time unix="` + strconv.Itoa(ent.UnixTime) + `" format="` + ent.DateTimeFormat + `">` + cntnt + "</tg-time>"
		}

		return prevText + `<tg-time unix="` + strconv.Itoa(ent.UnixTime) + `">` + cntnt + "</tg-time>"
	case botapi.EntityTextMention:
		return prevText + `<a href="tg://user?id=` + strconv.FormatInt(ent.User.ID, 10) + `">` + cntnt + "</a>"
	case "text_link": // Можно также использовать botapi.EntityTextLink, если библиотека поддерживает константу
		return prevText + `<a href="` + ent.URL + `">` + cntnt + "</a>"
	case botapi.EntityBlockquote:
		return prevText + `<blockquote>` + cntnt + "</blockquote>"
	case botapi.EntityExpandableBlockquote:
		return prevText + `<blockquote expandable>` + cntnt + "</blockquote>"
	default:
		return prevText + cntnt
	}
}

// closeHTMLTag makes sure to generate the correct HTML closing tag for a given opening tag.
func closeHTMLTag(s string) string {
	if !strings.HasPrefix(s, "span") {
		return s
	}

	return "span"
}

func writeFinalMarkdownV2(data []uint16, ent botapi.MessageEntity, start int, cntnt string) string {
	prevText := string(utf16.Decode(data[start:ent.Offset]))
	pre, cleanCntnt, post := splitEdgeWhitespace(cntnt, ent)

	switch ent.Type {
	case botapi.EntityBold, botapi.EntityItalic, botapi.EntityCode, botapi.EntityUnderline, botapi.EntityStrikethrough, botapi.EntitySpoiler:
		return prevText + pre + mdV2Map[ent.Type] + cleanCntnt + mdV2Map[ent.Type] + post
	case botapi.EntityPre:
		if ent.Language == "" {
			return prevText + pre + "```\n" + cleanCntnt + "```" + post
		}

		return prevText + pre + "```" + ent.Language + "\n" + cleanCntnt + "```" + post
	case botapi.EntityCustomEmoji:
		return prevText + pre + "![" + cleanCntnt + "](tg://emoji?id=" + ent.CustomEmojiID + ")" + post
	case botapi.EntityDateTime:
		if ent.DateTimeFormat != "" {
			return prevText + pre + "![" + cleanCntnt + "](tg://time?unix=" +
				strconv.Itoa(ent.UnixTime) + "&format=" + ent.DateTimeFormat + ")" + post
		}

		return prevText + pre + "![" + cleanCntnt + "](tg://time?unix=" + strconv.Itoa(ent.UnixTime) + ")" + post
	case botapi.EntityTextMention:
		return prevText + pre + "[" + cleanCntnt + "](tg://user?id=" + strconv.FormatInt(ent.User.ID, 10) + ")" + post
	case botapi.EntityTextLink:
		return prevText + pre + "[" + cleanCntnt + "](" + ent.URL + ")" + post
	case botapi.EntityBlockquote:
		return prevText + pre + ">" + strings.Join(strings.Split(cleanCntnt, "\n"), "\n>") + post
	case botapi.EntityExpandableBlockquote:
		return prevText + pre + "**>" + strings.Join(strings.Split(cleanCntnt, "\n"), "\n>") + "||" + post
	default:
		return prevText + cntnt
	}
}

func getUpperEntities(ents []botapi.MessageEntity) []botapi.MessageEntity {
	prev := 0
	uppers := make([]botapi.MessageEntity, 0, len(ents))

	for _, e := range ents {
		if e.Offset < prev {
			continue
		}

		uppers = append(uppers, e)
		prev = e.Offset + e.Length
	}

	return uppers
}

func getChildEntities(ent botapi.MessageEntity, ents []botapi.MessageEntity) []botapi.MessageEntity {
	end := ent.Offset + ent.Length
	children := make([]botapi.MessageEntity, 0, len(ents))

	for _, e := range ents {
		if e.Offset < ent.Offset || e == ent {
			continue
		}

		if e.Offset >= end {
			break
		}

		children = append(children, e)
	}

	return children
}

func splitEdgeWhitespace(text string, ent botapi.MessageEntity) (pre, cntnt, post string) {
	keepNewLines := ent.Type == botapi.EntityPre

	bd := strings.Builder{}
	rText := []rune(text)

	for i := 0; i < len(rText) && unicode.IsSpace(rText[i]) && (!keepNewLines || rText[i] != '\n'); i++ {
		bd.WriteRune(rText[i])
	}

	pre = bd.String()

	text = strings.TrimPrefix(text, pre)

	bd.Reset()

	for i := len(rText) - 1; i >= 0 && unicode.IsSpace(rText[i]); i-- {
		bd.WriteRune(rText[i])
	}

	post = bd.String()

	return pre, strings.TrimSuffix(text, post), post
}

func escapeContainedMDV1(data, mdType []rune) string {
	out := strings.Builder{}

	for _, x := range data {
		if slices.Contains(mdType, x) {
			out.WriteRune('\\')
		}

		out.WriteRune(x)
	}

	return out.String()
}
