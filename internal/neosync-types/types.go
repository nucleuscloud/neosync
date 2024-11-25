package neosynctypes

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type Version uint

const (
	V1            Version = iota + 1
	LatestVersion         = V1
)

const (
	NeosyncIntervalId = "NEOSYNC_INTERVAL"
	NeosyncArrayId    = "NEOSYNC_ARRAY"
)

type NeosyncAdapter interface {
	NeosyncMetadataType
	ScanPgx(value any) error
	ValuePgx() (any, error)
	ScanJson(value any) error
	ValueJson() (any, error)
}

type Neosync struct {
	Version Version `json:"version"`
	TypeId  string  `json:"type_id"`
}
type BaseType struct {
	Neosync Neosync `json:"_neosync"`
}

type NeosyncMetadataType interface {
	SetVersion(Version)
	GetVersion() Version
}

type NeosyncTypeOption func(NeosyncAdapter)

func WithVersion(version Version) NeosyncTypeOption {
	return func(t NeosyncAdapter) {
		if version == 0 {
			t.SetVersion(LatestVersion)
			return
		}
		t.SetVersion(version)
	}
}

type TypeRegistry struct {
	types map[string]map[Version]func() NeosyncAdapter
}

func NewTypeRegistry(logger *slog.Logger) *TypeRegistry {
	registry := &TypeRegistry{
		types: make(map[string]map[Version]func() NeosyncAdapter),
	}

	registry.Register(NeosyncIntervalId, LatestVersion, func() NeosyncAdapter {
		return NewInterval(WithVersion(LatestVersion))
	})

	registry.Register(NeosyncArrayId, LatestVersion, func() NeosyncAdapter {
		return NewNeosyncArray([]NeosyncAdapter{}, WithVersion(LatestVersion))
	})

	return registry
}

func (r *TypeRegistry) Register(typeId string, version Version, newTypeFunc func() NeosyncAdapter) {
	if _, exists := r.types[typeId]; !exists {
		r.types[typeId] = make(map[Version]func() NeosyncAdapter)
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
		return newTypeFunc(), nil
	}

	// Try LatestVersion
	if version != V1 {
		if newTypeFunc, ok := versionedTypes[LatestVersion]; ok {
			return newTypeFunc(), nil
		}
	}

	return nil, fmt.Errorf("unknown version %d for type ID: %s", version, typeId)
}

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
		// debug log
		return data, nil
	}

	typeId, ok := neosyncRaw["type_id"].(string)
	if !ok {
		// do something
		return data, nil
	}

	version, ok := neosyncRaw["version"].(Version)
	if !ok {
		version = LatestVersion
	}

	obj, err := r.New(typeId, version)
	if err != nil {
		return nil, err
	}

	// Handle arrays
	if typeId == NeosyncArrayId {
		elements, ok := rawMsg["elements"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array elements")
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
