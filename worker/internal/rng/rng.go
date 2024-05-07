package rng

import (
	"math/rand"
	"sync"
)

type Rand interface {
	Int63n(n int64) int64
	Intn(n int) int
	Float64() float64
}

var _ Rand = &Randomizer{} // Ensures Randomizer always conforms to the Rand interface

type Randomizer struct {
	rng *rand.Rand
	mu  sync.Mutex
}

func New(seed int64) *Randomizer {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec
	return &Randomizer{rng: rng}
}

func (r *Randomizer) Int63n(n int64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int63n(n)
}

func (r *Randomizer) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(n)
}

func (r *Randomizer) Float64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Float64()
}
