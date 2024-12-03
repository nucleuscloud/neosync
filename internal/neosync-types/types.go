package neosynctypes

import (
	"encoding/json"
	"fmt"
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

type NeosyncPgxValuer interface {
	ValuePgx() (any, error)
}

type NeosyncJsonValuer interface {
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
	setVersion(Version)
	GetVersion() Version
}

type NeosyncTypeOption func(NeosyncAdapter) error

func WithVersion(version Version) NeosyncTypeOption {
	return func(t NeosyncAdapter) error {
		if !IsValidVersion(version) {
			return fmt.Errorf("invalid Neosync Type version: %d", version)
		}
		if version == 0 {
			t.setVersion(LatestVersion)
			return nil
		}
		t.setVersion(version)
		return nil
	}
}

func applyOptions(t NeosyncAdapter, opts ...NeosyncTypeOption) error {
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return err
		}
	}
	return nil
}

func IsValidVersion(ver Version) bool {
	return ver == V1 || ver == LatestVersion
}

type JsonScanner struct{}

func (js *JsonScanner) ScanJson(value, target any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, target)
	case string:
		return json.Unmarshal([]byte(v), target)
	default:
		return fmt.Errorf("unsupported scan type for Json: %T", value)
	}
}

func (js *JsonScanner) ValueJson(value any) (any, error) {
	return json.Marshal(value)
}
