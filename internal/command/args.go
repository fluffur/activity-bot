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
	ArgTypeUser
)

type ArgRule struct {
	Name string
	Type ArgType
	Min  int
	Max  int
	// Variadic take whole string
	Variadic bool
}

func OneUserRule() ArgRule {
	return ArgRule{
		Name: "one_user",
		Type: ArgTypeUser,
		Min:  1,
		Max:  1,
	}
}

func OneNumberRule() ArgRule {
	return ArgRule{
		Name: "one_number",
		Type: ArgTypeNumber,
		Min:  1,
		Max:  1,
	}
}

func OneDateRule() ArgRule {
	return ArgRule{
		Name: "one_date",
		Type: ArgTypeDate,
		Min:  1,
		Max:  1,
	}
}

func OneTextRule() ArgRule {
	return ArgRule{
		Name: "one_text",
		Type: ArgTypeText,
		Min:  1,
		Max:  1,
	}
}
