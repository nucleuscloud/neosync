package neosynctypes

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type NeosyncTypeRegistry interface {
	Unmarshal(value any) (any, error)
}

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

	registry.Register(NeosyncBitsId, LatestVersion, func() (NeosyncAdapter, error) {
		return NewBits(WithVersion(LatestVersion))
	})

	registry.Register(NeosyncBinaryId, LatestVersion, func() (NeosyncAdapter, error) {
		return NewBinary(WithVersion(LatestVersion))
	})

	registry.Register(NeosyncDateTimeId, LatestVersion, func() (NeosyncAdapter, error) {
		return NewDateTime(WithVersion(LatestVersion))
	})

	return registry
}

func (r *TypeRegistry) Register(
	typeId string,
	version Version,
	newTypeFunc func() (NeosyncAdapter, error),
) {
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
	r.logger.Warn(
		fmt.Sprintf(
			"version %d not registered for Type Id: %s using latest version instead",
			version,
			typeId,
		),
	)
	if newTypeFunc, ok := versionedTypes[LatestVersion]; ok {
		return newTypeFunc()
	}

	return nil, fmt.Errorf(
		"unknown version %d for type Id: %s. latest version not found",
		version,
		typeId,
	)
}

// UnmarshalAny deserializes a value of type any into an appropriate type based on the Neosync type system.
// It handles specialized Neosync objects that contain type information in a "_neosync" metadata field.
//
// Parameters:
//   - value: any - The value to unmarshal, expected to be map[string]any
//
// Returns:
//   - any: The unmarshaled object, which could be:
//   - The original value if it's not a map[string]any
//   - A new instance of the appropriate type for Neosync objects
//   - A NeosyncArray containing unmarshaled elements for array types
func (r *TypeRegistry) Unmarshal(value any) (any, error) {
	rawMsg, ok := getMapFromAny(value)
	if !ok {
		return value, nil
	}

	neosyncRaw, ok := rawMsg["_neosync"].(map[string]any)
	if !ok {
		r.logger.Debug("value not a neosync type")
		return value, nil
	}

	typeId, ok := neosyncRaw["type_id"].(string)
	if !ok {
		r.logger.Debug("value missing _neosync.type_id field")
		return value, nil
	}

	var version Version
	if versionRaw, ok := neosyncRaw["version"].(float64); ok {
		version = Version(uint(versionRaw))
	} else {
		r.logger.Debug("value missing _neosync.version. Using latest version instead.")
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
			// Recursively unmarshal each element
			unmarshaledElement, err := r.Unmarshal(element)
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

	// Convert back to JSON to use standard unmarshaling
	data, err := json.Marshal(rawMsg)
	if err != nil {
		return nil, fmt.Errorf("error marshaling value: %w", err)
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, fmt.Errorf("error unmarshaling value into neosync types: %w", err)
	}

	return obj, nil
}

func getMapFromAny(value any) (map[string]any, bool) {
	// If value is already a map, return it
	if rawMsg, ok := value.(map[string]any); ok {
		return rawMsg, true
	}

	// If value is bytes, try to unmarshal into a map
	if rawBytes, ok := value.([]byte); ok && json.Valid(rawBytes) {
		var rawMsg map[string]any
		if err := json.Unmarshal(rawBytes, &rawMsg); err == nil {
			return rawMsg, true
		}
	}

	return nil, false
}
