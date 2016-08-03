package system

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type transaction struct {
	// TODO rename as unmerged
	jobs     map[*Unit]prospectiveJobs
	anchored map[*job]bool
	merged   set
}

type prospectiveJobs [JOB_TYPE_COUNT]*job

func newTransaction() (tr *transaction) {
	return &transaction{
		jobs:     map[*Unit]prospectiveJobs{},
		anchored: map[*job]bool{},
		merged:   set{},
	}
}

func (tr *transaction) Run() (err error) {
	var ordering []*job
	if ordering, err = tr.ordering(); err != nil {
		return
	}

	for _, j := range ordering {
		go j.Run()
	}
}

func (tr *transaction) ordering() (ordering []*job) {
	jobs := set{}

	// Do not stop until everything is merged
	for len(tr.jobs) > 0 {
		for u, prospective := range tr.jobs {
			var j *job

			if j, err = prospective.merge(); err != nil {
				for _, j := range prospective {
					for _, other := range prospective {
						if j == other || canMerge(j.typ, other.typ) {
							continue
						}

						switch {
						case j.anchored && other.anchored:
							return ErrDepConflict
						case !j.anchored && !other.anchored:
							// If there is an orphaned stop job - remove it
							// See https://goo.gl/z8SSDy
							switch {
							case j.typ == stop && len(j.conflictedBy) == 0:
								tr.delete(j)
							case other.typ == stop && len(other.conflictedBy) == 0:
								tr.delete(other)
							default:
								tr.delete(j)
							}
						case j.anchored:
							tr.delete(other)
						case other.anchored:
							tr.delete(j)
						}

					}
				}
				break
			}

			tr.merged.Put(j)
			delete(tr.jobs, u)
		}
	}

	return nil
}

func (j *job) isOrphan() bool {
	return 0 == len(j.wantedBy) == len(j.requiredBy) == len(j.conflictedBy)
}

// deletes j from transaction
// removes all references to j
// recurses on orphaned and broken jobs
func (tr *transaction) delete(j *job) {
	tr.jobs[j.unit][j.typ] = nil

	delete(tr.anchored, j)
	delete(tr.merged, j)

	for deps, f := range map[set]func(*job){
		j.wantedBy: func(depender *job) {
			delete(depender.wants, j)
		},
		j.requiredBy: func(depender *job) {
			delete(depender.requires, j)
			defer tr.delete(depender)
		},
		j.conflictedBy: func(depender *job) {
			delete(depender.conflicts, j)
			defer tr.delete(depender)
		},

		j.wants: func(dependency *job) {
			delete(dependency.wantedBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
		j.requires: func(dependency *job) {
			delete(dependency.requiredBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
		j.conflicts: func(dependency *job) {
			delete(dependency.conflictedBy, j)
			if dependency.isOrphan() {
				defer tr.delete(dependency)
			}
		},
	} {
		for dep := range deps {
			f(dep)
		}
	}
}

type mergeError struct {
	what, with *job
}

func newMergeErr(what, with *job) (me *mergeError) {
	return &mergeError{
		what: what,
		with: with,
	}
}

func canMerge(what, with jobType) (ok bool) {
	_, ok = mergeTable[what][with]
	return
}

func (prospective prospectiveJobs) merge() (merged *job, err error) {
	for _, j := range prospective {
		if j == nil {
			continue
		}

		if merged == nil {
			merged = j
			continue
		}

		var t jobType
		if t, ok = mergeTable[merged.typ][j.typ]; !ok {
			return nil, newMergeErr(merged, j)
		}

		merged.typ = t
	}

	return merged, nil
}

// recursively adds jobs to transaction
// tries to load dependencies not already present
func (tr *transaction) add(typ jobType, u *Unit, parent *job, matters, conflicts bool) (err error) {

	// TODO: decide if these checks are really necessary to do here,
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

	if isNew {
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

type graph struct {
	jobs     set
	visited  set
	ordering []*job
}

func newGraph(jobs set) (g *graph) {
	return &graph{
		jobs:     jobs,
		visited:  set{},
		ordering: make([]*job, 0, len(jobs)),
	}
}

func order(jobs set) (ordering []*job, err error) {
	log.Debugf("sys.order received jobs:\n%+v", jobs)

	g := newGraph(jobs)

	for j := range jobs {
		log.Debugf("Checking after of %s...", j)
		for _, depname := range j.unit.After() {
			var dep *Unit
			if dep, err = j.unit.system.Unit(depname); err != nil {
				continue
			}

			depJob, ok := jobs[dep]
			if ok {
				j.after.Put(depJob)
				depJob.before.Put(j)
			}
		}

		log.Debugf("Checking before of %s...", j)
		for _, depname := range j.unit.Before() {
			var dep *Unit
			if dep, err = j.unit.system.Unit(depname); err != nil {
				continue
			}

			depJob, ok := jobs[dep]
			if ok {
				depJob.after.Put(j)
				j.before.Put(depJob)
			}
		}
	}

	log.Debugf("starting DFS on graph:\n%+v", g)
	for j := range jobs {
		if err = g.traverse(j); err != nil {
			return nil, fmt.Errorf("Dependency cycle determined:\njob for %s depends on %s", j.unit.Name(), err)
		}
	}

	return g.ordering, nil
}

var errBlank = errors.New("")

func (g *graph) traverse(j *job) (err error) {
	log.Debugf("traverse %s\ngraph: %s\n\n", j, g)
	if g.ordering.Contains(j) {
		return nil
	}

	if g.visited.Contains(j) {
		return errBlank
	}

	g.visited.Put(j)

	for depJob := range j.after {
		if err = g.traverse(depJob); err != nil {
			if err == errBlank {
				return fmt.Errorf("%s\n", depJob)
			}
			return fmt.Errorf("%s\n%s depends on %s", j, j, err)
		}
	}

	g.visited.Remove(j)

	if !g.ordering.Contains(j) {
		g.ordering = append(g.ordering, j)
	}

	return nil
}
