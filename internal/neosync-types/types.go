package neosynctypes

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// scan from source to neosync type
// value from neosync type to destination

// source is postgres write to S3 pull from S3 write to Postgres in cli sync
// cli sync source is postgres convert to json bits stream to cli convert to pg types

// DataFormat represents supported data formats
type DataFormat int

const (
	Pgx DataFormat = iota
	Json
)

// Version represents the data format version
type Version uint

const (
	V1            Version = iota + 1
	LatestVersion         = V1
)

const (
	NeosyncInterval = "NEOSYNC_INTERVAL"
	NeosyncArray    = "NEOSYNC_ARRAY"
)

type PgxAdapter interface {
	ScanPgx(value any) error
	ValuePgx() (any, error)
}

type JsonAdapter interface {
	ScanJson(value any)
	ValueJson() ([]byte, error)
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

type NeosyncTypeOption[T NeosyncMetadataType] func(T)

func WithVersion[T NeosyncMetadataType](version Version) NeosyncTypeOption[T] {
	return func(t T) {
		if version == 0 {
			t.SetVersion(LatestVersion)
			return
		}
		t.SetVersion(version)
	}
}

type Interval struct {
	BaseType     `json:",inline"`
	Microseconds int64 `json:"microseconds"`
	Days         int32 `json:"days"`
	Months       int32 `json:"months"`
}

func (i *Interval) ScanPgx(value *pgtype.Interval) error {
	if value == nil || !value.Valid {
		return nil
	}
	i.Microseconds = value.Microseconds
	i.Days = value.Days
	i.Months = value.Months
	return nil
}

func (i *Interval) ValuePgx() (*pgtype.Interval, error) {
	return &pgtype.Interval{
		Microseconds: i.Microseconds,
		Days:         i.Days,
		Months:       i.Months,
		Valid:        true,
	}, nil
}

func (i *Interval) JsonScan(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, i)
	case string:
		return json.Unmarshal([]byte(v), i)
	default:
		return fmt.Errorf("unsupported scan type for JsonInterval: %T", value)
	}
}

func (i *Interval) JsonValue() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Interval) SetVersion(v Version) {
	i.Neosync.Version = v
}

func (i *Interval) GetVersion() Version {
	return i.Neosync.Version
}

func NewInterval(opts ...NeosyncTypeOption[*Interval]) *Interval {
	interval := &Interval{}
	interval.Neosync.TypeId = NeosyncInterval
	interval.SetVersion(LatestVersion)

	for _, opt := range opts {
		opt(interval)
	}
	return interval
}

type IntervalArray struct {
	*PgxArray[*Interval, *pgtype.Interval]
}

func NewIntervalArray(size int, intervalOpts []NeosyncTypeOption[*Interval], arrayOpts ...NeosyncTypeOption[*PgxArray[*Interval, *pgtype.Interval]]) *IntervalArray {
	return &IntervalArray{
		PgxArray: NewPgxArrayAdapter(size, NewInterval, intervalOpts, arrayOpts...),
	}
}

type PgxNeosyncAdapter[V any] interface {
	NeosyncMetadataType
	ScanPgx(value V) error
	ValuePgx() (V, error)
}

type JsonNeosyncAdapter interface {
	NeosyncMetadataType
	ScanJson(value any)
	ValueJson() ([]byte, error)
}

type PgxArray[T PgxNeosyncAdapter[V], V any] struct {
	BaseType `json:",inline"`
	Elements []T
}

func NewPgxArrayAdapter[T PgxNeosyncAdapter[V], V any](
	size int,
	intializer func(...NeosyncTypeOption[T]) T,
	initializerOpts []NeosyncTypeOption[T],
	arrayOpts ...NeosyncTypeOption[*PgxArray[T, V]],
) *PgxArray[T, V] {
	elements := make([]T, size)
	for i := range elements {
		elements[i] = intializer(initializerOpts...)
	}
	pgArray := &PgxArray[T, V]{
		Elements: elements,
	}
	pgArray.Neosync.TypeId = NeosyncArray
	pgArray.SetVersion(LatestVersion)

	for _, opt := range arrayOpts {
		opt(pgArray)
	}

	return &PgxArray[T, V]{
		Elements: elements,
	}
}

type PgxArrayAdapter interface {
	ScanArrayPgx(value any) error
	ValueArrayPgx() (any, error)
}

type JsonArrayAdapter interface {
	ScanArrayJson(value any) error
	ValueArrayJson() (any, error)
}

func (a *PgxArray[T, V]) SetVersion(v Version) {
	a.Neosync.Version = v
}

func (a *PgxArray[T, V]) GetVersion() Version {
	return a.Neosync.Version
}

func (a *PgxArray[T, V]) ScanArrayPgx(value any) error {
	valueSlice, ok := value.([]V)
	if !ok {
		return fmt.Errorf("expected []V, got %T", value)
	}
	if len(valueSlice) != len(a.Elements) {
		return fmt.Errorf("length mismatch: got %d elements, expected %d", len(valueSlice), len(a.Elements))
	}
	for i, v := range valueSlice {
		if err := a.Elements[i].ScanPgx(v); err != nil {
			return fmt.Errorf("scanning element %d: %w", i, err)
		}
	}
	return nil
}

func (a *PgxArray[T, V]) ValueArrayPgx() (any, error) {
	values := make([]any, len(a.Elements))
	for i, e := range a.Elements {
		v, err := e.ValuePgx()
		if err != nil {
			return nil, fmt.Errorf("getting value for element %d: %w", i, err)
		}
		values[i] = v
	}
	return values, nil
}

type TypeRegistry struct {
	types map[string]map[Version]func() any
}

func NewTypeRegistry() *TypeRegistry {
	registry := &TypeRegistry{
		types: make(map[string]map[Version]func() any),
	}

	registry.Register(NeosyncInterval, LatestVersion, func() any {
		return NewInterval(WithVersion[*Interval](LatestVersion))
	})

	registry.Register(NeosyncArray, LatestVersion, func() any {
		return NewPgxArrayAdapter(size, NewInterval, intervalOpts, arrayOpts...)
	})

	return registry
}

func (r *TypeRegistry) Register(typeId string, version Version, newTypeFunc func() any) {
	if _, exists := r.types[typeId]; !exists {
		r.types[typeId] = make(map[Version]func() any)
	}
	r.types[typeId][version] = newTypeFunc
}

func (r *TypeRegistry) New(typeId string, version Version) (any, error) {
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

// TODO add versioning here
func UnmarshalWithRegistry(data []byte, registry *TypeRegistry) (any, error) {
	var anyValue any
	if err := json.Unmarshal(data, &anyValue); err != nil {
		return nil, err
	}

	rawMsg, ok := anyValue.(map[string]any)
	if !ok {
		return anyValue, nil
	}

	neosyncRaw, ok := rawMsg["_neosync"].(map[string]any)
	if !ok {
		// debug log
		return rawMsg, nil
	}

	typeId, ok := neosyncRaw["type_id"].(string)
	if !ok {
		// do something
		return rawMsg, nil
	}

	version, ok := neosyncRaw["version"].(Version)
	if !ok {
		version = LatestVersion
	}

	obj, err := registry.New(typeId, version)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
