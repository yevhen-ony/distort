package resolve

import (
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	"errors"
	"fmt"
	"slices"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
	ErrMissingRaft  = errors.New("raft not set")
	ErrMissingSelf  = errors.New("self not set")
)

type Resolver struct {
	self t.MasterID 
	peers []t.MasterID
	c *Config
}

func newResolver(config *Config) *Resolver {
	self := t.MasterID(config.Self)
	peers := utils.Map(config.Peers, func(p string) t.MasterID {
		return t.MasterID(p)
	})
	return &Resolver{
		self: self,
		peers: peers,
		c: config,
	}
}

func New(config *Config) (*Resolver, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return newResolver(config), nil
}

func (r *Resolver) Refs() []t.MasterRef {
	return utils.Map(r.peers, r.makeRef)
}

func (r *Resolver) makeAPIAddr(peer t.MasterID) string {
	return fmt.Sprintf("%s:%d", peer, r.c.APIPort)
}

func (r *Resolver) makeRef(peer t.MasterID) t.MasterRef {
	return t.MasterRef {
		ID: peer,
		Addr: r.makeAPIAddr(peer),
	}
}

func (r *Resolver) Ref(id t.MasterID) (t.MasterRef, error) {
	if !slices.Contains(r.peers, id) {
		return t.MasterRef{}, ErrPeerNotFound
	}
	return r.makeRef(id), nil
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
	return r.makeRef(id), nil
}

func NewWithRaft(config *Config) (*Resolver, error) {
	if err := config.ValidateWithRaft(); err != nil {
		return nil, err
	}
	return newResolver(config), nil
}

type RaftRef struct {
	ID   t.MasterID
	Addr string
}

func (r *Resolver) makeRaftAddr(id t.MasterID) string {
	return fmt.Sprintf("%s:%d", id, r.c.RaftPort)
}

func (r *Resolver) makeRaftRef(id t.MasterID) RaftRef {
	return RaftRef{
		ID: id,
		Addr: r.makeRaftAddr(id),
	}
}

func (r *Resolver) RaftRef(id t.MasterID) (RaftRef, error) {
	if !slices.Contains(r.peers, id) {
		return RaftRef{}, ErrPeerNotFound
	}
	return r.makeRaftRef(id), nil
}

func (r *Resolver) RaftRefs() []RaftRef {
	return utils.Map(r.peers, r.makeRaftRef)
}

func (r *Resolver) RaftSelfRef() (RaftRef, error) {
	id, err := r.Self()
	if err != nil {
		return RaftRef{}, err 
	}
	return r.makeRaftRef(id), nil
}

