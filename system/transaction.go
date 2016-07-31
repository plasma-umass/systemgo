package system

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type transaction struct {
	jobs     map[*Unit][JOB_TYPE_COUNT]*job
	anchored map[*job]bool
	ordering []*job
}

func newTransaction() (tr *transaction) {
	return &transaction{
		jobMap:   jobMap{},
		anchored: map[*job]bool{},
	}
}

func (tr *transaction) Run() (err error) {
	if err = tr.loadDeps(); err != nil {
		return
	}

	var ordering []*job
	if ordering, err = tr.ordering(); err != nil {
		return
	}

	for _, j := range ordering {
		go j.Run()
	}
}

func (tr *transaction) add(typ jobType, u *Unit, parent *job, matters, conflicts) (err error) {

	// TODO: decide if these checks are really necessary to do here,
	// as they are performed exactly before a unit is started, reloaded etc.
	//
	//switch typ {
	//case reload:
	//	if !u.IsReloader() {
	//		return ErrNoReload
	//	}
	//case start:
	//	if !u.CanStart() {}
	//}

	if j = tr.jobs[u][typ]; j == nil {
		j = newJob(typ, u)
		jobs[u][typ] = j
		isNew = true
	}

	switch {
	case conflicts:
		parent.conflicts.Put(j)
		j.conflictedBy.Put(parent)
	case matters:
		parent.requires.Put(j)
		j.requiredBy.Put(parent)
	default:
		parent.wants.Put(j)
		j.wantedBy.Put(parent)
	}

	if isNew && recursive {
		for _, name := range u.Conflicts() {
			dep, err := u.System.Get(name)
			if err != nil {
				return
			}

			if err = tr.add(stop, dep, j, true, true); err != nil {
				return
			}
		}

		for _, name := range u.Requires() {
			dep, err := u.System.Get(name)
			if err != nil {
				return
			}

			if err = tr.add(start, dep, j, true, false); err != nil {
				return
			}
		}

		for _, name := range u.Wants() {
			dep, err := u.System.Get(name)
			if err != nil {
				continue
			}

			tr.add(start, dep, j, false, false)
		}
	}
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
		for _, depname := range u.Before() {
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
