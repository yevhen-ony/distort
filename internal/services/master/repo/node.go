package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"slices"
	"sync"
	"time"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)


type InMemNodeRegistry struct {
	nodes map[t.NodeID]*m.Node
	addrs map[string]t.NodeID

	mu sync.RWMutex
}

func NewInMemNodeRegistry() *InMemNodeRegistry {
	return &InMemNodeRegistry{
		nodes: map[t.NodeID]*m.Node{},
		addrs: map[string]t.NodeID{},
	}
}

func (r *InMemNodeRegistry) Register(_ context.Context, addr string) (t.NodeRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.addrs[addr]; ok {
		return t.NodeRef{}, m.ErrNodeAddrInUse 
	}
	
	nodeRef := t.NodeRef{
		ID: r.newNodeID(),
		Addr: addr,
	}
	r.addrs[nodeRef.Addr] = nodeRef.ID
	r.nodes[nodeRef.ID] = &m.Node{
		NodeRef: nodeRef,
		Stats: t.NodeStats{},
		LastSeenAt: time.Now().UTC(),
	}
	return nodeRef, nil 
}

func (r *InMemNodeRegistry) Unregister(_ context.Context, nid t.NodeID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	node := r.nodes[nid]
	if node == nil {
		return
	}

	delete(r.nodes, nid)
	delete(r.addrs, node.Addr)
}

func (r *InMemNodeRegistry) UpdateStats(_ context.Context, nid t.NodeID, stats t.NodeStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nid]
	if !ok {
		return m.ErrNodeNotFound	
	}

	node.Stats = stats
	node.LastSeenAt = time.Now().UTC()

	return nil 
}

func (r *InMemNodeRegistry) Get(ctx context.Context, nid t.NodeID) (m.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node := r.nodes[nid]
	if node == nil {
		return m.Node{}, m.ErrNodeNotFound	
	}
	return *node, nil
}

func (r *InMemNodeRegistry) GetMany(ctx context.Context, ids ...t.NodeID) []m.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]m.Node, 0, len(ids))
	for _, id := range ids {
		node := r.nodes[id]
		if node == nil {
			continue
		}
		nodes = append(nodes, *node)
	}
	return nodes
}

func (r *InMemNodeRegistry) Find(ctx context.Context, query m.NodeQuery) ([]m.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := []m.Node{}
	for _, node := range r.nodes {
		if query.MinFreeBytes > node.Stats.FreeBytes {
			continue
		}
		if slices.Contains(query.ExcludeIDs, node.NodeRef.ID) {
			continue
		}
		result = append(result, *node)
	}
	return result, nil
}

func (r *InMemNodeRegistry) newNodeID() t.NodeID {
	for {
		id := genNodeID()
		if _, ok := r.nodes[id]; !ok {
			return id	
		}
	}
}

func genNodeID() t.NodeID {
	var b [16]byte
	rand.Read(b[:])
	return t.NodeID(hex.EncodeToString(b[:]))
}

