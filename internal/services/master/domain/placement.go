package domain

import (
	"math/rand/v2"

	m "dos/internal/services/master"
)

type RandomPlacementPolicy struct{}

func (pp *RandomPlacementPolicy) Select(nodes []m.Node, n int) []m.Node {
	n = min(len(nodes), n)
	perm := rand.Perm(len(nodes))

	selected := make([]m.Node, 0, n)
	for i := range n {
		selected = append(selected, nodes[perm[i]])
	}
	return selected
}
