package benthosbuilder_connections

import (
	"fmt"
	"sync"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
)

// Global registry
var (
	defaultRegistry     *QueryBuilderRegistry
	defaultRegistryOnce sync.Once
)

// QueryBuilder interface
type QueryBuilder interface {
	BuildSelectQueryMap(
		driver string,
		tableFkConstraints map[string][]*sqlmanager_shared.ForeignConstraint,
		runConfigs []*tabledependency.RunConfig,
		subsetByForeignKeyConstraints bool,
		groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
	) (map[string]map[tabledependency.RunType]string, error)
}

// Function type that creates a QueryBuilder
type QueryBuilderFactory func() QueryBuilder

// Manages query builder registration and creation
type QueryBuilderRegistry struct {
	mu       sync.RWMutex
	builders map[string]QueryBuilderFactory
}

// Creates a new registry instance
func NewQueryBuilderRegistry() *QueryBuilderRegistry {
	return &QueryBuilderRegistry{
		builders: make(map[string]QueryBuilderFactory),
	}
}

// Adds a query builder factory for a specific connection
func (r *QueryBuilderRegistry) Register(driver string, factory QueryBuilderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders[driver] = factory
}

// Retrieves a query builder based on connection type
func (r *QueryBuilderRegistry) Get(driver string) (QueryBuilder, error) {
	r.mu.RLock()
	factory, exists := r.builders[driver]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no query builder registered for driver: %s", driver)
	}

	return factory(), nil
}

// Gets the default global QueryBuilderRegistry
func GetDefaultRegistry() *QueryBuilderRegistry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewQueryBuilderRegistry()
	})
	return defaultRegistry
}
