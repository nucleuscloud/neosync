package tablesync_shared

import (
	"context"
	"fmt"
	"sync"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"go.temporal.io/sdk/client"
	temporalclient "go.temporal.io/sdk/client"
)

const (
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

type GetNextIdentityBlock func(ctx context.Context, schema, table, column string, blockSize uint) (*IdentityRange, error)

type BlockAllocator interface {
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
	handle, err := i.temporalclient.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID: i.workflowId,
		RunID:      i.runId,
		UpdateName: AllocateIdentityBlock,
		Args: []any{&AllocateIdentityBlockRequest{
			Id:        token,
			BlockSize: blockSize,
		}},
		WaitForStage: client.WorkflowUpdateStageCompleted,
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
		maxAttempts := i.blockSize
		attempts := uint(0)

		// todo: ask AI if we should increase the max attempts due to possibility of randomly getting the same value every time
		// so that even if there is a value, we will still try to get a new one
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
	return 0, fmt.Errorf("unable to find unused value different from %d after getting new block", value)
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
