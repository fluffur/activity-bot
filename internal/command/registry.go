package command

import "slices"

type Registry struct {
	actions []*ActionDef

	byKey              map[string]*ActionDef
	byAlias            map[string]*ActionDef
	byCategory         map[Category][]*ActionDef
	commandsByCategory map[Category][]*ActionDef

	categories []Category
}

func NewRegistry() *Registry {
	return &Registry{
		byKey:              make(map[string]*ActionDef),
		byAlias:            make(map[string]*ActionDef),
		byCategory:         make(map[Category][]*ActionDef),
		commandsByCategory: make(map[Category][]*ActionDef),
	}
}

func (r *Registry) Add(defs []*ActionDef) {
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

func (r *Registry) All() []*ActionDef {
	return slices.Clone(r.actions)
}

func (r *Registry) Categories() []Category {
	return slices.Clone(r.categories)
}

func (r *Registry) ByCategory(category Category) []*ActionDef {
	return slices.Clone(r.byCategory[category])
}

func (r *Registry) Find(key string) (*ActionDef, bool) {
	def, ok := r.byKey[key]
	return def, ok
}

func (r *Registry) FindAlias(alias string) (*ActionDef, bool) {
	def, ok := r.byAlias[alias]
	return def, ok
}

func (r *Registry) FindByKeyOrAlias(name string) (*ActionDef, bool) {
	if def, ok := r.byKey[name]; ok {
		return def, true
	}

	def, ok := r.byAlias[name]

	return def, ok
}

func (r *Registry) Next(category Category, key string) *ActionDef {
	cmds := r.commandsByCategory[category]

	for i, cmd := range cmds {
		if cmd.Key == key && i+1 < len(cmds) {
			return cmds[i+1]
		}
	}

	return nil
}

func (r *Registry) Prev(category Category, key string) *ActionDef {
	cmds := r.commandsByCategory[category]

	for i, cmd := range cmds {
		if cmd.Key == key && i > 0 {
			return cmds[i-1]
		}
	}

	return nil
}
