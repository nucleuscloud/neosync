package neosynctypes

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type TypeRegistry struct {
	logger *slog.Logger
	types  map[string]map[Version]func() (NeosyncAdapter, error)
}

func NewTypeRegistry(logger *slog.Logger) *TypeRegistry {
	registry := &TypeRegistry{
		logger: logger,
		types:  make(map[string]map[Version]func() (NeosyncAdapter, error)),
	}

	registry.Register(NeosyncIntervalId, LatestVersion, func() (NeosyncAdapter, error) {
		return NewInterval(WithVersion(LatestVersion))
	})

	registry.Register(NeosyncArrayId, LatestVersion, func() (NeosyncAdapter, error) {
		return NewNeosyncArray([]NeosyncAdapter{}, WithVersion(LatestVersion))
	})

	return registry
}

func (r *TypeRegistry) Register(typeId string, version Version, newTypeFunc func() (NeosyncAdapter, error)) {
	if _, exists := r.types[typeId]; !exists {
		r.types[typeId] = make(map[Version]func() (NeosyncAdapter, error))
	}
	r.types[typeId][version] = newTypeFunc
}

func (r *TypeRegistry) New(typeId string, version Version) (NeosyncAdapter, error) {
	versionedTypes, ok := r.types[typeId]
	if !ok {
		return nil, fmt.Errorf("unknown type ID: %s", typeId)
	}

	// Try to get specific version
	if newTypeFunc, ok := versionedTypes[version]; ok {
		return newTypeFunc()
	}

	// Try LatestVersion
	r.logger.Warn(fmt.Sprintf("version %d not registered for Type Id: %s using latest version instead", version, typeId))
	if newTypeFunc, ok := versionedTypes[LatestVersion]; ok {
		return newTypeFunc()
	}

	return nil, fmt.Errorf("unknown version %d for type Id: %s. latest version not found.", version, typeId)
}

// Unmarshal deserializes JSON data into an appropriate type based on the Neosync type system.
// It handles both regular JSON data and specialized Neosync objects that contain type information
// in a "_neosync" metadata field.
//
// Parameters:
//   - data: []byte - Raw JSON data to unmarshal
//
// Returns:
//   - any: The unmarshaled object, which could be:
//   - The original data if it's not valid JSON
//   - A new instance of the appropriate type for Neosync objects
//   - A NeosyncArray containing unmarshaled elements for array types
//
// Special handling for arrays:
// When typeId is NeosyncArrayId, each element in the array is recursively
// unmarshaled and must implement the NeosyncAdapter interface.
func (r *TypeRegistry) Unmarshal(data []byte) (any, error) {
	isValidJson := json.Valid(data)
	if !isValidJson {
		return data, nil
	}
	var anyValue any
	if err := json.Unmarshal(data, &anyValue); err != nil {
		return nil, err
	}

	rawMsg, ok := anyValue.(map[string]any)
	if !ok {
		return data, nil
	}

	neosyncRaw, ok := rawMsg["_neosync"].(map[string]any)
	if !ok {
		r.logger.Debug("JSON data not a neosync type")
		return data, nil
	}

	typeId, ok := neosyncRaw["type_id"].(string)
	if !ok {
		r.logger.Debug("JSON data missing _neosync.type_id field")
		return data, nil
	}

	version, ok := neosyncRaw["version"].(Version)
	if !ok {
		r.logger.Debug("JSON data missing _neosync.version. Using latest version instead.")
		version = LatestVersion
	}
	r.logger.Debug(fmt.Sprintf("Neosync type %s version %d", typeId, version))

	obj, err := r.New(typeId, version)
	if err != nil {
		return nil, err
	}

	// Handle arrays
	if typeId == NeosyncArrayId {
		elements, ok := rawMsg["elements"].([]any)
		if !ok {
			return nil, fmt.Errorf("neosync array: invalid array elements")
		}

		neosyncArray := obj.(*NeosyncArray)
		neosyncArray.Elements = make([]NeosyncAdapter, len(elements))

		for i, element := range elements {
			elementBytes, err := json.Marshal(element)
			if err != nil {
				return nil, fmt.Errorf("error marshaling array element %d: %w", i, err)
			}

			// Recursively unmarshal each element
			unmarshaledElement, err := r.Unmarshal(elementBytes)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling array element %d: %w", i, err)
			}

			adapter, ok := unmarshaledElement.(NeosyncAdapter)
			if !ok {
				return nil, fmt.Errorf("array element %d is not a NeosyncAdapter", i)
			}

			neosyncArray.Elements[i] = adapter
		}
		return neosyncArray, nil
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
