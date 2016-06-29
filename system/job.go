package system

import "fmt"

type Job struct {
	Type   JobType
	Units  map[string]*Unit
	system *System
}

type JobType int

const (
	start JobType = iota
	stop
	isolate
)

func (sys *System) NewJob(typ JobType, names ...string) (j *Job, err error) {
	j = &Job{
		Type:   typ,
		Units:  map[string]*Unit{},
		system: sys,
	}

	for _, name := range names {
		if j.Units[name], err = sys.Get(name); err != nil {
			return nil, fmt.Errorf("Error loading %s: %s", name, err)
		}
	}

	//TODO: is stop case needed?
	switch j.Type {
	case start, isolate:
		for name, unit := range j.Units {
			if err = j.addDeps(unit); err != nil {
				return nil, fmt.Errorf("Error loading dependencies of %s: %s", name, err)
			}
		}
	}
	return
}

func (j *Job) Start() (err error) {
	ordering, err := j.Ordering()
	if err != nil {
		return
	}

	switch j.Type {
	case start:
		for _, u := range ordering {
			go u.Start()
		}
	case isolate:
		isolated := map[*Unit]struct{}{}

		for _, u := range ordering {
			go u.Start()
			isolated[u] = struct{}{}
		}

		for _, u := range j.system.units {
			if _, is := isolated[u]; !is {
				go u.Stop()
			}
		}
	case stop:
		for _, u := range ordering {
			go u.Stop()
		}
	}
	return
}
func (j *Job) Ordering() (ordering []*Unit, err error) {
	switch j.Type {
	case start, isolate:
		ordering, err = j.order()
	case stop: // TODO: Does stop also need a specific ordering?
		ordering = make([]*Unit, len(j.Units))
		for _, unit := range j.Units {
			ordering = append(ordering, unit)
		}
	}
	return
}

func (j *Job) order() (ordering []*Unit, err error) {
	g := &graph{
		map[*Unit]struct{}{},
		map[*Unit]struct{}{},
		map[*Unit]map[string]*Unit{},
		make([]*Unit, len(j.Units)),
	}

	ordering = g.ordering

	for name, unit := range j.Units {
		g.before[unit] = map[string]*Unit{}

		for _, depname := range unit.After() {
			if dep, ok := j.Units[depname]; ok {
				g.before[unit][depname] = dep
			}
		}

		for _, depname := range unit.Before() {
			if dep, ok := j.Units[depname]; ok {
				g.before[dep][name] = unit
			}
		}
	}

	for name, unit := range j.Units {
		if err = g.traverse(unit); err != nil {
			return nil, fmt.Errorf("Dependency cycle determined:\n%s depends on %s", name, err)
		}
	}

	return
}
func (j *Job) addDeps(u *Unit) (err error) {
	for _, name := range u.Requires() {
		if _, added := j.Units[name]; added {
			continue
		}

		var dep *Unit
		if dep, err = j.system.Get(name); err != nil {
			u.Log.Printf("Error loading %s: %s", name, err)
			continue
		}

		err = j.addDeps(dep)
	}

	if err != nil {
		return ErrDepFail
	}

	for _, name := range u.Wants() {
		var dep *Unit
		if dep, err = j.system.Get(name); err != nil {
			continue
		}

		j.addDeps(dep)
	}

	return
}

type graph struct {
	ordered  map[*Unit]struct{}
	visited  map[*Unit]struct{}
	before   map[*Unit]map[string]*Unit
	ordering []*Unit
}

func (g *graph) traverse(unit *Unit) (err error) {
	for name, dep := range g.before[unit] {
		if _, has := g.ordered[dep]; has {
			return
		}

		if _, has := g.visited[dep]; has {
			return fmt.Errorf("%s", name)
		}

		g.visited[dep] = struct{}{}

		if err = g.traverse(dep); err != nil {
			return fmt.Errorf("%s\n%s depends on %s", name, name, err)
		}

		delete(g.visited, dep)
	}

	g.ordering = append(g.ordering, unit)
	g.ordered[unit] = struct{}{}

	return
}
