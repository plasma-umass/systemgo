package system

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type transaction struct {
	jobs     map[*Unit]job
	anchor   *job
	ordering []*job
}

//func (sys *Daemon) newTransaction() (t *transaction) {
//	return &transaction{
//		system: sys,
//	}
//}
func newTransaction(anchor *job) (t *transaction) {
	return &transaction{
		jobs:   map[*Unit]*job{},
		anchor: anchor,
	}
}

func (t *transaction) Run() (err error) {
	for j := range t.ordering {
		go j.Run()
	}
}

func (u *Unit)dependencies() (deps set, err error) {
	var dep *Unit
}

func loadDep(u *Unit, deps set) (err error) {
	var dep *Unit
	for _, name := range u.Requires() {
		if dep, err = u.system.Get(name); err != nil {
			return
		}
		deps.Put(dep)
	}
}

// returns a map (name -> *Unit) containing dependencies and specified units
func loadDeps(units set) (deps set, err error) {
	var failed bool

	names := make([]string, len(units))
	for i, u := range units {
		names[i] = u.name
	}

	for len(names) > 0 {
		name := names[0]

		if !added(name) {
			var u *Unit
			if u, err = sys.Get(name); err != nil {
				return nil, fmt.Errorf("Error loading dependency: %s", name)
			}
			deps[name] = u

			names = append(names, u.Requires()...)

			for _, name := range u.Wants() {
				if !added(name) {
					deps[name], _ = sys.Get(name)
				}
			}
		}

		names = names[1:]
	}
	if failed {
		return nil, ErrDepFail
	}

	return
}

func (t *transaction) add(j *job) (err error) {
	//var u *Unit
	//if u, err = t.system.Get(name); err != nil {
	//return
	//}
	//assigned, has := t.jobs[u]
	assigned, has := t.jobs[j.unit]
	if !has {
		log.Debugf("Assigned a new job(%s) for %s", j, u.Name())
		t.jobs[j.unit] = j
		return
	}

	//return map[a][b] (nil)/(error)

	//if assigned != j {
	////return error
	//}
	switch {
	default:
		t.jobs[j.unit] = j
	}
	return
}

type set map[*Unit]struct{}

func (s set) Put(u *Unit) {
	s[u] = struct{}{}
}

func (s set) Contains(u *Unit) (ok bool) {
	_, ok = s[u]
	return
}

func (s set) Remove(u *Unit) {
	delete(s, u)
}

type ordering struct {
	indexes map[*Unit]int
	index   int
}

func newOrdering() (o *ordering) {
	return &ordering{
		indexes: map[*Unit]int{},
	}
}

func (o *ordering) Append(u *Unit) {
	o.indexes[u] = o.index
	o.index++
}

func (o *ordering) Contains(u *Unit) (ok bool) {
	_, ok = o.indexes[u]
	return
}

func (o *ordering) Ordering() (ordering []*Unit) {
	units := make([]*Unit, o.index)
	for u, i := range o.indexes {
		units[i] = u
	}

	ordering = make([]*Unit, len(o.indexes))
	for _, u := range units {
		ordering = append(ordering, u)
	}

	return
}

type graph struct {
	units   set
	visited set
	before  map[*Unit]set
	ordering
	// wantedBy []*Unit
	// requiredBy []*Unit

}

func newGraph(units set) (g *graph) {
	return &graph{
		units:    units,
		visited:  set{},
		before:   map[*Unit]set{},
		ordering: newOrdering(),
	}
}

func order(units set) (ordering []*Unit, err error) {
	log.Debugf("sys.order received units:\n%+v", units)

	g := newGraph(units)

	for u := range units {
		g.before[u] = set{}
	}

	for u := range units {
		log.Debugf("Checking after of %s...", u.name)
		for _, depname := range u.After() {
			log.Debugf("%s after %s", u.name, depname)
			if dep, err := u.system.Unit(depname); err == nil && units.Contains(dep) {
				g.before[u].Put(dep)
			}
		}

		log.Debugf("Checking before of %s...", u.name)
		for _, depname := range unit.Before() {
			log.Debugf("%s before %s", u.name, depname)
			if dep, err := u.system.Unit(depname); err == nil && units.Contains(dep) {
				g.before[dep].Put(u)
			}
		}
	}

	log.Debugf("starting DFS on graph:\n%+v", g)
	for u := range units {
		log.Debugf("picking %s", u.name)
		if err = g.traverse(u); err != nil {
			return nil, fmt.Errorf("Dependency cycle determined:\n%s depends on %s", u.name, err)
		}
	}

	return g.ordering.Order(), nil
}

var errBlank = errors.New("")

func (g *graph) traverse(u *Unit) (err error) {
	log.Debugf("traverse %p, graph %p, ordering:\n%v", u, g, g.ordering)
	if g.ordering.Contains(u) {
		return nil
	}

	if g.visited.Contains(u) {
		return errBlank
	}

	g.visited.Put(u)

	for name, dep := range g.before[u] {
		if err = g.traverse(dep); err != nil {
			if err == errBlank {
				return fmt.Errorf("%s\n", name)
			}
			return fmt.Errorf("%s\n%s depends on %s", name, name, err)
		}
	}

	g.visited.Remove(u)

	if !g.ordering.Contains(u) {
		log.Debugf()
		g.ordering.Append(u)
	}

	return nil
}
