package rng

import (
	"math/rand/v2"
	"sync"
)

type Rand interface {
	Int63n(n int64) int64
	Intn(n int) int
	Float64() float64
	Int63() int64
	Int() int
	Uint() uint
}

var _ Rand = &Randomizer{} // Ensures Randomizer always conforms to the Rand interface

type Randomizer struct {
	rng *rand.Rand
	mu  sync.Mutex
}

func New(seed int64) *Randomizer {
	return NewSplit(uint64(seed), uint64(seed)) //nolint:gosec
}

func NewSplit(seed1, seed2 uint64) *Randomizer {
	rng := rand.New(rand.NewPCG(seed1, seed2)) //nolint:gosec
	return &Randomizer{rng: rng}
}

func (r *Randomizer) Int63n(n int64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int64N(n)
}

func (r *Randomizer) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.IntN(n)
}

func (r *Randomizer) Float64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Float64()
}

func (r *Randomizer) Int63() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int64()
}

func (r *Randomizer) Int() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Int()
}

func (r *Randomizer) Uint() uint {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Uint()
}
