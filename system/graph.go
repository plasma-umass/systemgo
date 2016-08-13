package system

import (
	"errors"
	"fmt"

	"github.com/apex/log"
)

type graph struct {
	visited, ordered set
	ordering         []*job
}

func newGraph() (g *graph) {
	return &graph{
		visited: set{},
		ordered: set{},
	}
}

var errBlank = errors.New("")

func (g *graph) String() (out string) {
	for _, j := range g.ordering {
		out = fmt.Sprintf("%s->%s", out, j)
	}
	return
}

func (g *graph) order(j *job) (err error) {
	log.WithField("func", "order").Debugf("called %s\ngraph: %s\n\n", j, g)
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
				return fmt.Errorf("%s\n", depJob)
			}
			return fmt.Errorf("%s\n%s depends on %s", j, j, err)
		}
	}

	delete(g.visited, j)

	if !g.ordered.Contains(j) {
		g.ordering = append(g.ordering, j)
		g.ordered.Put(j)
	}

	return nil
}
