package clientmanager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	temporalclient "go.temporal.io/sdk/client"
)

// ClientCache manages temporal client instances with thread-safe caching
type ClientCache struct {
	mu sync.RWMutex
	// Map of config hash to cached clients
	clients map[string]*cachedClient
}

type cachedClient struct {
	workflowClient  temporalclient.Client
	namespaceClient temporalclient.NamespaceClient
	scheduleClient  temporalclient.ScheduleClient
	referenceCount  int
	config          *TemporalConfig
}

func NewClientCache() *ClientCache {
	return &ClientCache{
		clients: make(map[string]*cachedClient),
	}
}

// getOrCreateClient retrieves a cached client or creates a new one
func (c *ClientCache) getOrCreateClient(
	ctx context.Context,
	config *TemporalConfig,
	factory ClientFactory,
	logger *slog.Logger,
) (*clientHandle, error) {
	hash := config.Hash()

	c.mu.Lock()
	defer c.mu.Unlock()

	if client, exists := c.clients[hash]; exists {
		client.referenceCount++
		return &clientHandle{
			cache:  c,
			client: client,
			config: config,
		}, nil
	}
	workflowClient, err := factory.CreateWorkflowClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow client: %w", err)
	}

	namespaceClient, err := factory.CreateNamespaceClient(ctx, config, logger)
	if err != nil {
		workflowClient.Close()
		return nil, fmt.Errorf("failed to create namespace client: %w", err)
	}

	scheduleClient := workflowClient.ScheduleClient()

	client := &cachedClient{
		workflowClient:  workflowClient,
		namespaceClient: namespaceClient,
		scheduleClient:  scheduleClient,
		referenceCount:  1,
		config:          config,
	}
	c.clients[hash] = client

	return &clientHandle{
		cache:  c,
		client: client,
		config: config,
	}, nil
}

// releaseClient decrements reference count and cleans up if no more references
func (c *ClientCache) releaseClient(config *TemporalConfig) {
	hash := config.Hash()

	c.mu.Lock()
	defer c.mu.Unlock()

	client, exists := c.clients[hash]
	if !exists {
		return
	}

	client.referenceCount--
	// not deleting the temproal connection because it has trouble re-opening
	// if client.referenceCount <= 0 {
	// 	client.workflowClient.Close()
	// 	client.namespaceClient.Close()
	// 	delete(c.clients, hash)
	// }
}

type clientHandle struct {
	cache  *ClientCache
	client *cachedClient
	config *TemporalConfig
}

func (h *clientHandle) WorkflowClient() temporalclient.Client {
	return h.client.workflowClient
}

func (h *clientHandle) NamespaceClient() temporalclient.NamespaceClient {
	return h.client.namespaceClient
}

func (h *clientHandle) Release() {
	h.cache.releaseClient(h.config)
}
