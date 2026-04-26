package domain

import (
	m "dos/internal/services/master"
	t "dos/internal/common/types"
)


func toNodeRef(nodes ...m.Node) []t.NodeRef {
	if nodes == nil {
		return nil
	}
	result := make([]t.NodeRef, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, t.NodeRef{
			ID: node.ID,
			Addr: node.Stats.Addr,
		})
	}
	return result
}
