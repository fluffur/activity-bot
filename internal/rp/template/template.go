package rptemplate

import (
	"activity-bot/internal/model"
	"regexp"
	"strings"
)

var (
	actorGenderVariantRe  = regexp.MustCompile(`\{actor:([^{}]+)\}`)
	targetGenderVariantRe = regexp.MustCompile(`\{target:([^{}]+)\}`)
	pairGenderVariantRe   = regexp.MustCompile(`\{pair:([^{}]+)\}`)
	actorBracketRe        = regexp.MustCompile(`([\p{L}\-]+)\(а\)`)
	bracketVariantRe      = regexp.MustCompile(`[\p{L}\-]+\(а\)`)
)

var templateAliasReplacer = strings.NewReplacer(
	"{A}", "{actor}",
	"{B}", "{target}",
	"{a}", "{actor}",
	"{b}", "{target}",
	"{A:", "{actor:",
	"{B:", "{target:",
	"{a:", "{actor:",
	"{b:", "{target:",
)

var templateStartMarkers = []string{
	"{actor", "{target", "{command", "{pair", "{A", "{B", "{a", "{b",
}

// Normalize converts user-friendly RP syntax to canonical placeholders.
func Normalize(raw string) string {
	normalized := templateAliasReplacer.Replace(raw)
	return actorBracketRe.ReplaceAllString(normalized, "{actor:${1}|${1}а}")
}

// ResolveVariants resolves placeholder variants by actor/target genders.
func ResolveVariants(raw, actorGender, targetGender string) string {
	actorKey := genderKey(actorGender)
	targetKey := genderKey(targetGender)

	replaced := actorGenderVariantRe.ReplaceAllStringFunc(raw, func(match string) string {
		return pickVariant(match, actorGenderVariantRe, actorKey, "")
	})
	replaced = targetGenderVariantRe.ReplaceAllStringFunc(replaced, func(match string) string {
		return pickVariant(match, targetGenderVariantRe, targetKey, "")
	})
	replaced = pairGenderVariantRe.ReplaceAllStringFunc(replaced, func(match string) string {
		return pickVariant(match, pairGenderVariantRe, actorKey, targetKey)
	})

	return replaced
}

// SplitTriggerAndTemplate separates RP trigger text from template part.
func SplitTriggerAndTemplate(raw string) (string, string) {
	templateStart := len(raw)
	for _, marker := range templateStartMarkers {
		if idx := strings.Index(raw, marker); idx >= 0 && idx < templateStart {
			templateStart = idx
		}
	}
	if loc := bracketVariantRe.FindStringIndex(raw); loc != nil && loc[0] < templateStart {
		templateStart = loc[0]
	}
	if templateStart == len(raw) {
		return strings.TrimSpace(raw), ""
	}
	return strings.TrimSpace(raw[:templateStart]), strings.TrimSpace(raw[templateStart:])
}

func pickVariant(raw string, re *regexp.Regexp, firstKey, secondKey string) string {
	groups := re.FindStringSubmatch(raw)
	if len(groups) != 2 {
		return raw
	}

	variants := strings.Split(groups[1], "|")
	for i := range variants {
		variants[i] = strings.TrimSpace(variants[i])
	}

	switch re {
	case actorGenderVariantRe, targetGenderVariantRe:
		if len(variants) != 2 {
			return raw
		}
		if firstKey == "f" {
			return variants[1]
		}
		return variants[0]
	case pairGenderVariantRe:
		if len(variants) != 4 {
			return raw
		}
		switch firstKey + secondKey {
		case "mm":
			return variants[0]
		case "mf":
			return variants[1]
		case "fm":
			return variants[2]
		case "ff":
			return variants[3]
		default:
			return variants[0]
		}
	default:
		return raw
	}
}

func genderKey(gender string) string {
	if gender == model.GenderFemale {
		return "f"
	}
	return "m"
}
