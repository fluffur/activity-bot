package rule

type RuleType string

const (
	RuleUser               RuleType = "user"
	RuleNumber             RuleType = "number"
	RuleDuration           RuleType = "duration"
	RuleDateTime           RuleType = "datetime"
	RuleDateTimeOrDuration RuleType = "duration_or_datetime"
	RuleText               RuleType = "text"
)

const RuleVariadic = -1

type TextValidator func(string) bool

type Rule struct {
	Type         RuleType
	IsOptional   bool
	CountArgs    int
	OnNextRow    bool
	TextValidate TextValidator
}

func User() Rule {
	return Rule{
		Type:      RuleUser,
		CountArgs: 1,
	}
}

func Number() Rule {
	return Rule{
		Type:      RuleNumber,
		CountArgs: 1,
	}
}

func DateTimeOrDuration() Rule {
	return Rule{
		Type:      RuleDateTimeOrDuration,
		CountArgs: 1,
	}
}

func Text() Rule {
	return Rule{
		Type:      RuleText,
		CountArgs: 1,
	}
}

func (r Rule) Optional() Rule {
	r.IsOptional = true
	return r
}

func (r Rule) Required() Rule {
	r.IsOptional = false
	return r
}

func (r Rule) Variadic() Rule {
	r.CountArgs = RuleVariadic
	return r
}

func (r Rule) Count(n int) Rule {
	r.CountArgs = n
	return r
}

func (r Rule) Validate(fn TextValidator) Rule {
	r.TextValidate = fn
	return r
}

func Duration() Rule {
	return Rule{
		Type:      RuleDuration,
		CountArgs: 1,
	}
}
