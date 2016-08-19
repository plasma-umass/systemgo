package system

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type transaction struct {
	unmerged map[*Unit]*prospectiveJobs
	merged   map[*Unit]*job
}

type prospectiveJobs struct {
	anchored, optional [job_type_count]*job
}

func newTransaction() (tr *transaction) {
	log.Debugf("newTransaction")

	return &transaction{
		unmerged: map[*Unit]*prospectiveJobs{},
		merged:   map[*Unit]*job{},
	}
}

func (tr *transaction) Run() (err error) {
	log.WithField("transaction", tr).Debugf("tr.Run")

	if err = tr.merge(); err != nil {
		return
	}

	var ordering []*job
	if ordering, err = tr.order(); err != nil {
		return
	}

	for _, j := range ordering {
		if j.IsRedundant() {
			continue
		}

		log.Debugf("dispatching job for %s", j.unit.Name())
		go j.Run()
	}
	return
}

// recursively adds jobs to transaction
// tries to load dependencies not already present
func (tr *transaction) add(typ jobType, u *Unit, parent *job, required, anchor bool) (err error) {
	log.WithFields(log.Fields{
		"typ":      typ,
		"u":        u,
		"parent":   parent,
		"required": required,
		"anchor":   anchor,
	}).Debug("tr.add")
	// TODO: decide if these checks are necessary to do here,
	// as they are performed by the unit method calls already
	//
	//switch typ {
	//case reload:
	//	if !u.IsReloader() {
	//		return ErrNoReload
	//	}
	//case start:
	//	if !u.CanStart() {}
	//}
	var j *job
	var isNew bool

	if tr.unmerged[u] == nil {
		tr.unmerged[u] = &prospectiveJobs{}
	}

	if anchor {
		if j = tr.unmerged[u].anchored[typ]; j == nil {
			j = newJob(typ, u)
			log.Debugf("Created %s", j)

			tr.unmerged[u].anchored[typ] = j

			isNew = true
		}
	} else {
		if j = tr.unmerged[u].optional[typ]; j == nil {
			j = newJob(typ, u)
			log.Debugf("Created %s", j)

			tr.unmerged[u].optional[typ] = j

			isNew = true
		}
	}

	if parent != nil {
		if required {
			parent.requires.Put(j)
			j.requiredBy.Put(parent)
		} else {
			parent.wants.Put(j)
			j.wantedBy.Put(parent)
		}
	}

	if isNew && typ != stop {
		for _, name := range u.Conflicts() {
			dep, err := u.System.Get(name)
			if err != nil {
				return err
			}

			if err = tr.add(stop, dep, j, true, anchor); err != nil {
				return err
			}
		}

		for _, name := range u.Requires() {
			dep, err := u.System.Get(name)
			if err != nil {
				return err
			}

			if err = tr.add(start, dep, j, true, anchor); err != nil {
				return err
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

	return nil
}

func (tr *transaction) merge() (err error) {
	log.Debug("tr.merge")

	for u, prospective := range tr.unmerged {
		var merged *job

		for _, j := range prospective.anchored {
			if j == nil {
				continue
			}

			if merged == nil {
				merged = j
			} else {
				if err = merged.mergeWith(j); err != nil {
					return
				}
			}
		}

		for _, j := range prospective.optional {
			if j == nil {
				continue
			}

			if merged == nil {
				merged = j
			} else {
				if err = merged.mergeWith(j); err != nil {
					// TODO be smart when deleting unmergeable jobs
					tr.delete(j)
				}
			}

			prospective.optional[j.typ] = nil
		}

		tr.merged[u] = merged
		delete(tr.unmerged, u)
	}

	return nil
}

// TODO implement something along these lines
//for _, j := range prospective {
//	for _, other := range prospective {
//		if j == other || canMerge(j.typ, other.typ) {
//			continue
//		}

//		switch {
//		case tr.anchored[j] && tr.anchored[other]:
//			return ErrDepConflict
//		case !tr.anchored[j] && !tr.anchored[other]:
//			// If there is an orphaned stop job - remove it
//			// See https://goo.gl/z8SSDy
//			switch {
//			case j.typ == stop && len(j.conflictedBy) == 0:
//				tr.delete(j)
//			case other.typ == stop && len(other.conflictedBy) == 0:
//				tr.delete(other)
//			default:
//				tr.delete(j)
//			}
//		case tr.anchored[j]:
//			tr.delete(other)
//		case tr.anchored[other]:
//			tr.delete(j)
//		}

//	}
//}
//break
//}

// deletes j from transaction
// removes all references to j
// recurses on orphaned and broken jobs
func (tr *transaction) delete(j *job) {
	log.WithField("j", j).Debug("tr.delete")

	delete(tr.merged, j.unit)

	for deps, f := range map[*set]func(*job){
		&j.wantedBy: func(depender *job) {
			delete(depender.wants, j)
		},
		&j.requiredBy: func(depender *job) {
			delete(depender.requires, j)
			defer tr.delete(depender)
		},
		&j.conflictedBy: func(depender *job) {
			delete(depender.conflicts, j)
			defer tr.delete(depender)
		},

		&j.wants: func(dependency *job) {
			delete(dependency.wantedBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
		&j.requires: func(dependency *job) {
			delete(dependency.requiredBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
		&j.conflicts: func(dependency *job) {
			delete(dependency.conflictedBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
	} {
		for dep := range *deps {
			f(dep)
		}
	}
}

func canMerge(what, with jobType) (ok bool) {
	_, ok = mergeTable[what][with]
	return
}

func (tr *transaction) order() (ordering []*job, err error) {
	log.Debug("tr.order")

	g := newGraph()

	for u, j := range tr.merged {
		if j.typ == stop {
			// TODO Introduce stop job ordering(if needed)
			continue
		}

		log.Debugf("Checking after of %s...", j.unit.Name())
		for _, depname := range u.After() {
			var dep *Unit
			if dep, err = u.System.Unit(depname); err != nil {
				continue
			}

			depJob, ok := tr.merged[dep]
			if ok {
				j.after.Put(depJob)
				depJob.before.Put(j)
			}
		}

		log.Debugf("Checking before of %s...", j.unit.Name())
		for _, depname := range u.Before() {
			var dep *Unit
			if dep, err = u.System.Unit(depname); err != nil {
				continue
			}

			depJob, ok := tr.merged[dep]
			if ok {
				depJob.after.Put(j)
				j.before.Put(depJob)
			}
		}
	}

	g.ordering = make([]*job, 0, len(tr.merged))
	for _, j := range tr.merged {
		if err = g.order(j); err != nil {
			return nil, fmt.Errorf("Dependency cycle determined:\njob for %s depends on %s", j.unit.Name(), err)
		}
	}

	return g.ordering, nil
}

type graph struct {
	visited, ordered set
	ordering         []*job
}

func newGraph() (g *graph) {
	log.Debugf("newGraph")

	return &graph{
		visited: set{},
		ordered: set{},
	}
}

var errBlank = errors.New("")

func (g *graph) order(j *job) (err error) {
	log.WithField("j", j).Debugf("g.order")

	if g.ordered.Contains(j) {
		return nil
	}

	if g.visited.Contains(j) {
		return errBlank
	}

	g.visited.Put(j)

	for depJob := range j.after {
		if err = g.order(depJob); err != nil {
			if err == errBlank {
				return fmt.Errorf("%s\n", depJob.unit.Name())
			}
			return fmt.Errorf("%v\n%s depends on %s", j.unit.Name(), j.unit.Name(), err)
		}
	}

	delete(g.visited, j)

	if !g.ordered.Contains(j) {
		g.ordering = append(g.ordering, j)
		g.ordered.Put(j)
	}

	return nil
}
