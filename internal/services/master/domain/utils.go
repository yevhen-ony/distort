package domain

import (
	m "dos/internal/services/master"
)


func toNodeAccess(nodes ...m.Node) []m.NodeAccess {
	if nodes == nil {
		return nil
	}
	result := make([]m.NodeAccess, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, m.NodeAccess{
			NodeID: node.ID,
			Addr: node.Report.Addr,
		})
	}
	return result
}
