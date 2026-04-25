package domain

import (
	m "dos/internal/services/master"
	t "dos/internal/common/types"
)


func toNodeAccess(nodes ...m.Node) []t.NodeAccess {
	if nodes == nil {
		return nil
	}
	result := make([]t.NodeAccess, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, t.NodeAccess{
			NodeID: node.ID,
			Addr: node.Report.Addr,
		})
	}
	return result
}
