package transformers

import (
	"context"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIdentityAllocator implements the IdentityAllocator interface for testing
type MockIdentityAllocator struct {
	identities map[uint]uint
}

func (m *MockIdentityAllocator) GetIdentity(ctx context.Context, token string, value *uint) (uint, error) {
	if value == nil {
		return 1, nil
	}
	if val, ok := m.identities[*value]; ok {
		return val, nil
	}
	return *value + 1, nil
}

// MockZeroIdentityAllocator always returns 0 for testing error cases
type MockZeroIdentityAllocator struct{}

func (m *MockZeroIdentityAllocator) GetIdentity(ctx context.Context, token string, value *uint) (uint, error) {
	return 0, nil
}

func Test_RegisterTransformIdentityScramble(t *testing.T) {
	t.Run("successfully registers function", func(t *testing.T) {
		env := bloblang.NewEnvironment()
		allocator := &MockIdentityAllocator{
			identities: make(map[uint]uint),
		}
		err := RegisterTransformIdentityScramble(env, allocator)
		require.NoError(t, err)

		// Test the function can be executed through bloblang
		ex, err := env.Parse(`root = transform_identity_scramble(value: 1, token: "test-token")`)
		require.NoError(t, err)

		res, err := ex.Query(nil)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func Test_transformIdentityScramble(t *testing.T) {
	allocator := &MockIdentityAllocator{
		identities: map[uint]uint{
			1: 100,
			2: 200,
		},
	}

	t.Run("handles nil value", func(t *testing.T) {
		result, err := transformIdentityScramble(allocator, "test-token", nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles int value", func(t *testing.T) {
		result, err := transformIdentityScramble(allocator, "test-token", int(1))
		require.NoError(t, err)
		assert.Equal(t, uint(100), result)
	})

	t.Run("handles int64 value", func(t *testing.T) {
		result, err := transformIdentityScramble(allocator, "test-token", int64(2))
		require.NoError(t, err)
		assert.Equal(t, uint(200), result)
	})

	t.Run("handles uint value", func(t *testing.T) {
		result, err := transformIdentityScramble(allocator, "test-token", uint(1))
		require.NoError(t, err)
		assert.Equal(t, uint(100), result)
	})

	t.Run("handles negative int value", func(t *testing.T) {
		result, err := transformIdentityScramble(allocator, "test-token", int(-1))
		require.NoError(t, err)
		assert.Equal(t, uint(1), result) // Should get identity for 0
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		_, err := transformIdentityScramble(allocator, "test-token", "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to get identity from value as input was string")
	})

	t.Run("rejects zero identity value", func(t *testing.T) {
		zeroAllocator := &MockZeroIdentityAllocator{}
		_, err := transformIdentityScramble(zeroAllocator, "test-token", uint(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to get identity from value as generated identity was 0")
	})
}
