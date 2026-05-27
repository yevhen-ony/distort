package resolve

import t "dos/internal/common/types"

type ResolverWithRaft struct {
	*Resolver
}

func NewWithRaft(config *Config) (*ResolverWithRaft, error) {
	if err := config.ValidateWithRaft(); err != nil {
		return nil, err
	}
	resolver := &ResolverWithRaft{
		Resolver: newResolver(config),
	}
	return resolver, nil
}

type RaftRef struct {
	ID   t.MasterID
	Addr string
}

func (r *ResolverWithRaft) RaftRefs() []RaftRef {
	peers := make([]RaftRef, 0, len(r.peers))
	for _, p := range r.peers {
		peers = append(peers, RaftRef{ID: p.ID, Addr: p.RaftAddr})
	}
	return peers
}

func (r *ResolverWithRaft) RaftRef(id t.MasterID) (RaftRef, error) {
	p, ok := r.peers[id]
	if !ok {
		return RaftRef{}, ErrPeerNotFound
	}
	ref := RaftRef{ID: p.ID, Addr: p.RaftAddr}
	return ref, nil 
}

func (r *ResolverWithRaft) RaftSelfRef() (RaftRef, error) {
	id, err := r.Self()
	if err != nil {
		return RaftRef{}, err 
	}
	return r.RaftRef(id)
}

