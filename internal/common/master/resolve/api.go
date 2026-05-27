package resolve

import (
	t "dos/internal/common/types"
	"errors"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
	ErrMissingRaft  = errors.New("raft not set")
	ErrMissingSelf  = errors.New("self not set")
)

type Resolver struct {
	self  t.MasterID
	peers map[t.MasterID]Peer
}

func newResolver(config *Config) *Resolver {
	peers := make(map[t.MasterID]Peer, len(config.Peers))
	for _, peer := range config.Peers {
		peers[peer.ID] = peer
	}

	return &Resolver{
		self:  config.Self,
		peers: peers,
	}
}

func New(config *Config) (*Resolver, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newResolver(config), nil
}

func (r *Resolver) Refs() []t.MasterRef {
	peers := make([]t.MasterRef, 0, len(r.peers))
	for _, p := range r.peers {
		peers = append(peers, t.MasterRef{ID: p.ID, Addr: p.APIAddr})
	}
	return peers
}

func (r *Resolver) Ref(id t.MasterID) (t.MasterRef, error) {
	p, ok := r.peers[id]
	if !ok {
		return t.MasterRef{}, ErrPeerNotFound
	}
	ref := t.MasterRef{ID: p.ID, Addr: p.APIAddr}
	return ref, nil
}

func (r *Resolver) Self() (t.MasterID, error) {
	if r.self == "" {
		return "", ErrMissingSelf
	}
	return r.self, nil
}

func (r *Resolver) SelfRef() (t.MasterRef, error) {
	id, err := r.Self()	
	if err != nil {
		return t.MasterRef{}, err
	}
	return r.Ref(id)
}
