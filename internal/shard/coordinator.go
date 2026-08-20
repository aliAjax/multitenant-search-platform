package shard

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"hash/fnv"
	"sync"
	"time"
)

type State string

const (
	StatePrimary     State = "primary"
	StateReplica     State = "replica"
	StateFailed      State = "failed"
	StateRebalancing State = "rebalancing"
)

type Assignment struct {
	CollectionID string    `json:"collection_id"`
	Shard        int       `json:"shard"`
	Node         string    `json:"node"`
	State        State     `json:"state"`
	Epoch        int64     `json:"epoch"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Repository interface {
	Put(context.Context, Assignment) error
	List(context.Context, string) []Assignment
}
type Coordinator struct {
	repo  Repository
	clock platform.Clock
	mu    sync.RWMutex
	nodes map[string]bool
}

func NewCoordinator(r Repository, c platform.Clock) *Coordinator {
	return &Coordinator{repo: r, clock: c, nodes: map[string]bool{}}
}
func (c *Coordinator) AddNode(id string)    { c.mu.Lock(); c.nodes[id] = true; c.mu.Unlock() }
func (c *Coordinator) RemoveNode(id string) { c.mu.Lock(); delete(c.nodes, id); c.mu.Unlock() }
func (c *Coordinator) Assign(ctx context.Context, col string, shards int) error {
	c.mu.RLock()
	nodes := []string{}
	for n := range c.nodes {
		nodes = append(nodes, n)
	}
	c.mu.RUnlock()
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes: %w", platform.ErrConflict)
	}
	for i := 0; i < shards; i++ {
		a := Assignment{CollectionID: col, Shard: i, Node: nodes[i%len(nodes)], State: StatePrimary, Epoch: time.Now().UnixNano(), UpdatedAt: c.clock.Now()}
		if e := c.repo.Put(ctx, a); e != nil {
			return fmt.Errorf("assign shard: %w", e)
		}
	}
	return nil
}
func (c *Coordinator) Route(collection, doc string, shards int) int {
	if shards <= 1 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(collection + "\x00" + doc))
	return int(h.Sum32() % uint32(shards))
}
func (c *Coordinator) Failover(ctx context.Context, col string, shard int, node string) error {
	for _, a := range c.repo.List(ctx, col) {
		if a.Shard == shard {
			a.Node = node
			a.State = StatePrimary
			a.Epoch++
			a.UpdatedAt = c.clock.Now()
			return c.repo.Put(ctx, a)
		}
	}
	return platform.ErrNotFound
}
