package javascript_processor

import (
	"sync"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBatchBenthosValueApi(t *testing.T) {
	api := newBatchBenthosValueApi()
	assert.NotNil(t, api)
	assert.Nil(t, api.message)
}

func TestSetAndGetMessage(t *testing.T) {
	api := newBatchBenthosValueApi()
	msg := service.NewMessage([]byte("test message"))

	api.SetMessage(msg)
	retrievedMsg := api.Message()

	assert.Equal(t, msg, retrievedMsg)
}

func TestSetAndGetBytes(t *testing.T) {
	api := newBatchBenthosValueApi()
	msg := service.NewMessage([]byte("original"))
	api.SetMessage(msg)

	// Test SetBytes
	newBytes := []byte("updated content")
	api.SetBytes(newBytes)

	// Test AsBytes
	retrievedBytes, err := api.AsBytes()
	require.NoError(t, err)
	assert.Equal(t, newBytes, retrievedBytes)
}

func TestSetAndGetStructured(t *testing.T) {
	api := newBatchBenthosValueApi()
	msg := service.NewMessage(nil)
	api.SetMessage(msg)

	// Test SetStructured
	structData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	api.SetStructured(structData)

	// Test AsStructured
	retrievedData, err := api.AsStructured()
	require.NoError(t, err)
	assert.Equal(t, structData, retrievedData)
}

func TestMetaGetAndSet(t *testing.T) {
	api := newBatchBenthosValueApi()
	msg := service.NewMessage(nil)
	api.SetMessage(msg)

	// Test MetaSetMut
	api.MetaSetMut("testKey", "testValue")

	// Test MetaGet
	value, exists := api.MetaGet("testKey")
	assert.True(t, exists)
	assert.Equal(t, "testValue", value)

	// Test non-existent key
	_, exists = api.MetaGet("nonExistentKey")
	assert.False(t, exists)
}

func TestConcurrentAccess(t *testing.T) {
	api := newBatchBenthosValueApi()
	msg := service.NewMessage([]byte("original"))
	api.SetMessage(msg)

	var wg sync.WaitGroup
	concurrentOperations := 100

	// Test concurrent reads
	wg.Add(concurrentOperations)
	for i := 0; i < concurrentOperations; i++ {
		go func() {
			defer wg.Done()
			_, err := api.AsBytes()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// Test concurrent writes
	wg.Add(concurrentOperations)
	for i := 0; i < concurrentOperations; i++ {
		go func(i int) {
			defer wg.Done()
			api.SetBytes([]byte("updated content"))
		}(i)
	}
	wg.Wait()

	// Test mixed reads and writes
	wg.Add(concurrentOperations * 2)
	for i := 0; i < concurrentOperations; i++ {
		go func() {
			defer wg.Done()
			api.SetBytes([]byte("concurrent write"))
		}()
		go func() {
			defer wg.Done()
			_, _ = api.AsBytes()
		}()
	}
	wg.Wait()

	// Test concurrent metadata operations
	wg.Add(concurrentOperations * 2)
	for i := 0; i < concurrentOperations; i++ {
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i%10))
			api.MetaSetMut(key, i)
		}(i)
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i%10))
			_, _ = api.MetaGet(key)
		}(i)
	}
	wg.Wait()
}

func TestNilMessageHandling(t *testing.T) {
	api := newBatchBenthosValueApi()
	// No message set, should handle gracefully

	// These should not panic but may return errors
	bytes, err := api.AsBytes()
	assert.Error(t, err)
	assert.Nil(t, bytes)
	assert.Contains(t, err.Error(), "message is nil")

	structured, err := api.AsStructured()
	assert.Error(t, err)
	assert.Nil(t, structured)
	assert.Contains(t, err.Error(), "message is nil")

	// These should not panic and just be no-ops
	api.SetBytes([]byte("test"))
	api.SetStructured(map[string]interface{}{"key": "value"})
	api.MetaSetMut("key", "value")

	// This should return false for exists
	val, exists := api.MetaGet("key")
	assert.False(t, exists)
	assert.Nil(t, val)

	// Now set a message and verify operations work
	msg := service.NewMessage([]byte("test"))
	api.SetMessage(msg)

	bytes, err = api.AsBytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), bytes)
}
