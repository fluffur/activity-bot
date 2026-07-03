package command

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"
)

type TriggerType string

const (
	TriggerCommand  TriggerType = "command"
	TriggerCallback TriggerType = "callback"
)

type Category string

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
	Rules      []predicate.Rule
}

type Registry struct {
	cmds       []*ActionDef
	idx        map[string]*ActionDef
	categories []Category
}

func NewRegistry() *Registry {
	return &Registry{idx: make(map[string]*ActionDef)}
}

func (r *Registry) AddCategory(category Category) {
	r.categories = append(r.categories, category)
}

func (r *Registry) Categories() []Category {
	return r.categories
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

func (r *Registry) Grouped() map[Category][]*ActionDef {
	out := make(map[Category][]*ActionDef)
	for _, c := range r.cmds {
		out[c.Category] = append(out[c.Category], c)
	}
	return out
}

func (r *Registry) ByCategory(category Category) []*ActionDef {
	var out []*ActionDef

	for _, cmd := range r.cmds {
		if cmd.Category == category {
			out = append(out, cmd)
		}
	}

	return out
}

func (r *Registry) Find(key string) (*ActionDef, bool) {
	for _, cmd := range r.cmds {
		if cmd.Key == key {
			return cmd, true
		}
	}

	return nil, false
}

func (r *Registry) CategoryIndex(category Category) int {
	for i, c := range r.categories {
		if c == category {
			return i
		}
	}

	return -1
}

func (r *Registry) CommandIndex(category Category, key string) int {
	cmds := r.ByCategory(category)

	for i, cmd := range cmds {
		if cmd.Key == key {
			return i
		}
	}

	return -1
}

func (r *Registry) NextCommand(category Category, key string) *ActionDef {
	cmds := r.ByCategory(category)

	for i, cmd := range cmds {
		if cmd.Key == key && i+1 < len(cmds) {
			return cmds[i+1]
		}
	}

	return nil
}

func (r *Registry) PrevCommand(category Category, key string) *ActionDef {
	cmds := r.ByCategory(category)

	for i, cmd := range cmds {
		if cmd.Key == key && i > 0 {
			return cmds[i-1]
		}
	}

	return nil
}

func (r *Registry) FindByKeyOrAlias(name string) (*ActionDef, bool) {
	for _, cmd := range r.cmds {
		if cmd.Key == name {
			return cmd, true
		}
	}

	for _, cmd := range r.cmds {
		for _, alias := range cmd.Aliases {
			if alias == name {
				return cmd, true
			}
		}
	}

	return nil, false
}
