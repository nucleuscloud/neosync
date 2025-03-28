package tablesync_shared

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	temporalclient "go.temporal.io/sdk/client"
)

const (
	// Method name for the Temporal Update handler that will allocate a block of identities
	AllocateIdentityBlock = "allocate-identity-block"
)

type AllocateIdentityBlockRequest struct {
	Id        string // This will be the value present in the ColumnIdentityCursors map (key)
	BlockSize uint   // The size of the block the caller wishes to allocate
}
type AllocateIdentityBlockResponse struct {
	StartValue uint // Inclusive
	EndValue   uint // Exclusive - represents the next value after the last valid value
}

type IdentityRange struct {
	StartValue uint // Inclusive
	EndValue   uint // Exclusive - represents the next value after the last valid value
}

type BlockAllocator interface {
	// GetNextBlock returns the next block of identities for the given token
	GetNextBlock(ctx context.Context, token string, blockSize uint) (*IdentityRange, error)
}

type TemporalBlockAllocator struct {
	temporalclient temporalclient.Client
	workflowId     string
	runId          string
}

func NewTemporalBlockAllocator(temporalclient temporalclient.Client, workflowId, runId string) *TemporalBlockAllocator {
	return &TemporalBlockAllocator{
		temporalclient: temporalclient,
		workflowId:     workflowId,
		runId:          runId,
	}
}

func (i *TemporalBlockAllocator) GetNextBlock(ctx context.Context, token string, blockSize uint) (*IdentityRange, error) {
	handle, err := i.temporalclient.UpdateWorkflow(ctx, temporalclient.UpdateWorkflowOptions{
		WorkflowID: i.workflowId,
		RunID:      i.runId,
		UpdateName: AllocateIdentityBlock,
		Args: []any{&AllocateIdentityBlockRequest{
			Id:        token,
			BlockSize: blockSize,
		}},
		WaitForStage: temporalclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send update to get next block size for identity %s: %w", token, err)
	}
	var resp *AllocateIdentityBlockResponse
	err = handle.Get(ctx, &resp)
	if err != nil {
		return nil, fmt.Errorf("unable to get next block size for identity %s: %w", token, err)
	}
	return &IdentityRange{
		StartValue: resp.StartValue,
		EndValue:   resp.EndValue,
	}, nil
}

type IdentityAllocator interface {
	// Given an (optional) input value, will return a value that is not the same as the input value
	GetIdentity(ctx context.Context, token string, value *uint) (uint, error)
}

type MultiIdentityAllocator struct {
	blockAllocator BlockAllocator
	blockSize      uint
	seed           uint64

	mu *sync.Mutex

	allocators map[string]*SingleIdentityAllocator
}

func NewMultiIdentityAllocator(blockAllocator BlockAllocator, blockSize uint, seed uint64) *MultiIdentityAllocator {
	return &MultiIdentityAllocator{
		blockAllocator: blockAllocator,
		blockSize:      blockSize,
		seed:           seed,
		allocators:     make(map[string]*SingleIdentityAllocator),
		mu:             &sync.Mutex{},
	}
}

func (i *MultiIdentityAllocator) GetIdentity(ctx context.Context, token string, value *uint) (uint, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	allocator, ok := i.allocators[token]
	if !ok {
		allocator = NewSingleIdentityAllocator(i.blockAllocator, i.blockSize, rng.NewSplit(i.seed, hashToSeed(token)))
		i.allocators[token] = allocator
	}
	return allocator.GetIdentity(ctx, token, value)
}

func hashToSeed(token string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(token))
	return h.Sum64()
}

// Note: This allocator caches the used values in a memory-map.
// This could be problematic if the worker restarts and is retried.
// This will result in Temporal returning the same block for allocation, but the allocator will no longer have
// the cache of values used in the previous attempt.
// To mitigate this, we need to store the used values in a durable store (e.g. Redis)
type SingleIdentityAllocator struct {
	blockAllocator BlockAllocator
	blockSize      uint
	rand           rng.Rand

	mu           *sync.Mutex
	currentBlock *IdentityRange
	usedValues   map[uint]struct{}
}

func NewSingleIdentityAllocator(blockAllocator BlockAllocator, blockSize uint, rand rng.Rand) *SingleIdentityAllocator {
	return &SingleIdentityAllocator{
		blockAllocator: blockAllocator,
		blockSize:      blockSize,
		rand:           rand,
		mu:             &sync.Mutex{},
	}
}

func (i *SingleIdentityAllocator) GetIdentity(ctx context.Context, token string, value *uint) (uint, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Helper function for the main retry logic
	tryGetValue := func() (uint, bool) {
		// Increase max attempts to reduce the probability of missing an available value
		maxAttempts := i.blockSize * 2 // Double the attempts to be more thorough
		attempts := uint(0)

		for attempts < maxAttempts {
			randomValue := i.currentBlock.StartValue + i.rand.Uint()%(i.currentBlock.EndValue-i.currentBlock.StartValue)
			if value == nil || *value != randomValue {
				if _, used := i.usedValues[randomValue]; !used {
					i.usedValues[randomValue] = struct{}{}
					return randomValue, true
				}
			}
			attempts++
		}
		return 0, false
	}

	// Initial block setup if needed
	if i.currentBlock == nil {
		block, err := i.blockAllocator.GetNextBlock(ctx, token, i.blockSize)
		if err != nil {
			return 0, fmt.Errorf("unable to get next block for identity %s: %w", token, err)
		}
		i.currentBlock = block
		i.usedValues = make(map[uint]struct{})
	}

	// Check if we've exhausted the current block
	if uint(len(i.usedValues)) >= i.blockSize {
		block, err := i.blockAllocator.GetNextBlock(ctx, token, i.blockSize)
		if err != nil {
			return 0, fmt.Errorf("unable to get next block for identity %s: %w", token, err)
		}
		i.currentBlock = block
		i.usedValues = make(map[uint]struct{})
	}

	// Try with current block
	if val, ok := tryGetValue(); ok {
		return val, nil
	}

	// If we couldn't find a value, get a new block and try again
	block, err := i.blockAllocator.GetNextBlock(ctx, token, i.blockSize)
	if err != nil {
		return 0, fmt.Errorf("unable to get next block for identity %s: %w", token, err)
	}
	i.currentBlock = block
	i.usedValues = make(map[uint]struct{})

	// Try with new block
	if val, ok := tryGetValue(); ok {
		return val, nil
	}

	// If we still can't find a value, something is seriously wrong
	return 0, errors.New("unable to find unused value different from input after getting new block")
}

// handles allocating blocks of integers to be used for auto increment columns
type IdentityCursor struct {
	CurrentValue uint
}

func NewDefaultIdentityCursor() *IdentityCursor {
	return &IdentityCursor{
		CurrentValue: 0,
	}
}
