package command

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
)

type TriggerType string

const (
	TriggerCommand  TriggerType = "command"
	TriggerCallback TriggerType = "callback"
)

type Category string

const (
	CategoryHelp       Category = "help"
	CategoryStats      Category = "stats"
	CategorySummon     Category = "summon"
	CategoryModeration Category = "moderation"
	CategoryEvents     Category = "events"
)

func Categories() []Category {
	return []Category{
		CategoryHelp,
		CategorySummon,
		CategoryEvents,
	}
}

type Scope int

const (
	ScopeAny     Scope = 0
	ScopePrivate Scope = 1
	ScopeGroup   Scope = 2
)

type ActionDef struct {
	Key string

	Aliases []string

	Trigger TriggerType

	Parent *ActionDef

	MinStatus   chatmember.Status
	Category    Category
	Description i18n.MessageID
	Examples    []i18n.MessageID
	Scope       Scope

	ShowInHelp bool
}

type Registry struct {
	cmds []*ActionDef
	idx  map[string]*ActionDef
}

func NewRegistry() *Registry {
	return &Registry{idx: make(map[string]*ActionDef)}
}

func (r *Registry) Add(def *ActionDef) {
	r.cmds = append(r.cmds, def)
	for _, a := range def.Aliases {
		r.idx[a] = def
	}
}

func (r *Registry) Get(alias string) (*ActionDef, bool) {
	def, ok := r.idx[alias]
	return def, ok
}

func (r *Registry) All() []*ActionDef {
	out := make([]*ActionDef, len(r.cmds))
	copy(out, r.cmds)
	return out
}

func (r *Registry) ByCategory() map[Category][]*ActionDef {
	out := make(map[Category][]*ActionDef)
	for _, c := range r.cmds {
		out[c.Category] = append(out[c.Category], c)
	}
	return out
}
