package command

import "slices"

type Registry struct {
	actions []*Action

	byKey              map[string]*Action
	byAlias            map[string]*Action
	byCategory         map[Category][]*Action
	commandsByCategory map[Category][]*Action

	categories []Category
}

func NewRegistry() *Registry {
	return &Registry{
		byKey:              make(map[string]*Action),
		byAlias:            make(map[string]*Action),
		byCategory:         make(map[Category][]*Action),
		commandsByCategory: make(map[Category][]*Action),
	}
}

func (r *Registry) Add(defs []*Action) {
	for _, def := range defs {
		r.actions = append(r.actions, def)
		r.byKey[def.Key] = def

		for _, key := range def.Trigger.IndexKeys() {
			r.byAlias[key] = def
		}

		if _, ok := r.byCategory[def.Category]; !ok {
			r.categories = append(r.categories, def.Category)
		}

		r.byCategory[def.Category] = append(r.byCategory[def.Category], def)

		if _, ok := def.Trigger.(*CommandTrigger); ok {
			r.commandsByCategory[def.Category] = append(
				r.commandsByCategory[def.Category],
				def,
			)
		}
	}
}

func (r *Registry) All() []*Action {
	return slices.Clone(r.actions)
}

func (r *Registry) Categories() []Category {
	return slices.Clone(r.categories)
}

func (r *Registry) ByCategory(category Category) []*Action {
	return slices.Clone(r.byCategory[category])
}

func (r *Registry) Find(key string) (*Action, bool) {
	def, ok := r.byKey[key]
	return def, ok
}

func (r *Registry) FindAlias(alias string) (*Action, bool) {
	def, ok := r.byAlias[alias]
	return def, ok
}

func (r *Registry) FindByKeyOrAlias(name string) (*Action, bool) {
	if def, ok := r.byKey[name]; ok {
		return def, true
	}

	def, ok := r.byAlias[name]

	return def, ok
}

func (r *Registry) Next(category Category, key string) *Action {
	cmds := r.commandsByCategory[category]

	for i, cmd := range cmds {
		if cmd.Key == key && i+1 < len(cmds) {
			return cmds[i+1]
		}
	}

	return nil
}

func (r *Registry) Prev(category Category, key string) *Action {
	cmds := r.commandsByCategory[category]

	for i, cmd := range cmds {
		if cmd.Key == key && i > 0 {
			return cmds[i-1]
		}
	}

	return nil
}
