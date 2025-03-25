package tablesync_shared

import (
	"context"
	"fmt"
	"testing"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_SingleIdentityAllocator_GetIdentity(t *testing.T) {
	t.Run("returns new identity when no input value provided", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)

		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(10)).
			Return(&IdentityRange{StartValue: 1, EndValue: 11}, nil).
			Once()

		randGen.EXPECT().
			Uint().
			Return(uint(5)).
			Once()

		allocator := NewSingleIdentityAllocator(blockAllocator, 10, randGen)
		result, err := allocator.GetIdentity(context.Background(), "test-token", nil)

		require.NoError(t, err)
		require.Equal(t, uint(6), result) // StartValue(1) + random(5) = 6
	})

	t.Run("returns different identity when input value provided", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)
		inputVal := uint(6)

		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(10)).
			Return(&IdentityRange{StartValue: 1, EndValue: 11}, nil).
			Once()

		// First try returns same as input, second try returns different
		randGen.EXPECT().
			Uint().
			Return(uint(5)).
			Once() // This will generate 6, which matches input
		randGen.EXPECT().
			Uint().
			Return(uint(7)).
			Once() // This will generate 8, which is different

		allocator := NewSingleIdentityAllocator(blockAllocator, 10, randGen)
		result, err := allocator.GetIdentity(context.Background(), "test-token", &inputVal)

		require.NoError(t, err)
		require.Equal(t, uint(8), result)
		require.NotEqual(t, inputVal, result)
	})

	t.Run("gets new block when current block is exhausted", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)

		// First block
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 1, EndValue: 3}, nil).
			Once()

		// Second block when first is exhausted
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 3, EndValue: 5}, nil).
			Once()

		// Generate values for first block
		randGen.EXPECT().Uint().Return(uint(0)).Once() // Will generate 1
		randGen.EXPECT().Uint().Return(uint(1)).Once() // Will generate 2

		// Generate value for second block
		randGen.EXPECT().Uint().Return(uint(0)).Once() // Will generate 3

		allocator := NewSingleIdentityAllocator(blockAllocator, 2, randGen)

		// First allocation
		result1, err := allocator.GetIdentity(context.Background(), "test-token", nil)
		require.NoError(t, err)
		require.Equal(t, uint(1), result1)

		// Second allocation
		result2, err := allocator.GetIdentity(context.Background(), "test-token", nil)
		require.NoError(t, err)
		require.Equal(t, uint(2), result2)

		// Third allocation (should get new block)
		result3, err := allocator.GetIdentity(context.Background(), "test-token", nil)
		require.NoError(t, err)
		require.Equal(t, uint(3), result3)
	})

	t.Run("returns error when block allocator fails", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)

		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(10)).
			Return(nil, fmt.Errorf("allocation failed")).
			Once()

		allocator := NewSingleIdentityAllocator(blockAllocator, 10, randGen)
		_, err := allocator.GetIdentity(context.Background(), "test-token", nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to get next block for identity test-token")
	})

	t.Run("gets new block when all values in current block conflict with input", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)
		inputVal := uint(1)

		// First block (small block where all values would conflict)
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 1, EndValue: 3}, nil).
			Once()

		// Second block when first block has no valid values
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 10, EndValue: 12}, nil).
			Once()

		// First block attempts (will try blockSize*2=4 times)
		randGen.EXPECT().Uint().Return(uint(0)).Times(4) // Will generate 1 (matches input) four times

		// Second block attempts
		randGen.EXPECT().Uint().Return(uint(0)).Once() // Will generate 10

		allocator := NewSingleIdentityAllocator(blockAllocator, 2, randGen)

		// Try to get a value - should fail in first block and succeed in second
		result, err := allocator.GetIdentity(context.Background(), "test-token", &inputVal)
		require.NoError(t, err)
		require.Equal(t, uint(10), result)
	})

	t.Run("returns error when unable to find value in new block", func(t *testing.T) {
		blockAllocator := NewMockBlockAllocator(t)
		randGen := rng.NewMockRand(t)
		inputVal := uint(1)

		// First block
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 1, EndValue: 3}, nil).
			Once()

		// Second block
		blockAllocator.EXPECT().
			GetNextBlock(mock.Anything, "test-token", uint(2)).
			Return(&IdentityRange{StartValue: 1, EndValue: 3}, nil).
			Once()

		// First block attempts (will try blockSize*2=4 times)
		randGen.EXPECT().Uint().Return(uint(0)).Times(4) // Will generate 1 (matches input) four times

		// Second block attempts (will try blockSize*2=4 times)
		randGen.EXPECT().Uint().Return(uint(0)).Times(4) // Will generate 1 (matches input) four times

		allocator := NewSingleIdentityAllocator(blockAllocator, 2, randGen)

		_, err := allocator.GetIdentity(context.Background(), "test-token", &inputVal)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to find unused value different from input")
	})
}
