package command

type Offset struct {
	Start, End int
}

const MaxAny = -1

type ArgType int

const (
	ArgTypeText ArgType = iota
	ArgTypeNumber
	ArgTypeDate
	ArgTypeAnyUser
	ArgTypeOnlyUserSender
	ArgTypeMentionedUser
)

type ArgRule struct {
	Name string
	Type ArgType
	Min  int
	Max  int
	// Variadic take whole string
	Variadic bool
}

func AnyUserRule() ArgRule {
	return ArgRule{
		Name: "any_user",
		Type: ArgTypeAnyUser,
		Min:  1,
		Max:  1,
	}
}

func MentionedUserRule() ArgRule {
	return ArgRule{
		Name: "mentioned_user",
		Type: ArgTypeMentionedUser,
		Min:  1,
		Max:  1,
	}
}

func NumberRule() ArgRule {
	return ArgRule{
		Name: "one_number",
		Type: ArgTypeNumber,
		Min:  1,
		Max:  1,
	}
}

func (r ArgRule) SetRange(min int, max int) ArgRule {
	r.Min = min
	r.Max = max

	return r
}

func (r ArgRule) SetVariadic(variadic bool) ArgRule {
	r.Variadic = variadic

	return r
}

func OneDateRule() ArgRule {
	return ArgRule{
		Name: "one_date",
		Type: ArgTypeDate,
		Min:  1,
		Max:  1,
	}
}

func TextRule() ArgRule {
	return ArgRule{
		Name: "one_text",
		Type: ArgTypeText,
		Min:  1,
		Max:  1,
	}
}

func OptionalVariadicText() ArgRule {
	return ArgRule{
		Name:     "optional_variadic_text",
		Type:     ArgTypeText,
		Min:      0,
		Max:      1,
		Variadic: true,
	}
}

func OptionalDateRangeRule() ArgRule {
	return ArgRule{
		Name: "period",
		Type: ArgTypeDate,
		Min:  0,
		Max:  2,
	}
}
