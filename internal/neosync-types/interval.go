package neosynctypes

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Interval struct {
	BaseType     `json:",inline"`
	Microseconds int64 `json:"microseconds"`
	Days         int32 `json:"days"`
	Months       int32 `json:"months"`
}

func (i *Interval) ScanPgx(value any) error {
	if value == nil {
		return nil
	}
	interval, ok := value.(*pgtype.Interval)
	if !ok {
		return fmt.Errorf("expected *pgtype.Interval, got %T", value)
	}
	if !interval.Valid {
		return nil
	}
	i.Microseconds = interval.Microseconds
	i.Days = interval.Days
	i.Months = interval.Months
	return nil
}

func (i *Interval) ValuePgx() (any, error) {
	return &pgtype.Interval{
		Microseconds: i.Microseconds,
		Days:         i.Days,
		Months:       i.Months,
		Valid:        true,
	}, nil
}

func (i *Interval) ScanJson(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, i)
	case string:
		return json.Unmarshal([]byte(v), i)
	default:
		return fmt.Errorf("unsupported scan type for JsonInterval: %T", value)
	}
}

func (i *Interval) ValueJson() (any, error) {
	return json.Marshal(i)
}

func (i *Interval) SetVersion(v Version) {
	i.Neosync.Version = v
}

func (i *Interval) GetVersion() Version {
	return i.Neosync.Version
}

func NewIntervalFromPgx(value any, opts ...NeosyncTypeOption) (*Interval, error) {
	interval := NewInterval(opts...)
	err := interval.ScanPgx(value)
	if err != nil {
		return nil, err
	}
	return interval, nil
}

func NewInterval(opts ...NeosyncTypeOption) *Interval {
	interval := &Interval{}
	interval.Neosync.TypeId = NeosyncIntervalId
	interval.SetVersion(LatestVersion)

	for _, opt := range opts {
		opt(interval)
	}
	return interval
}

func NewIntervalArrayFromPgx(elements []*pgtype.Interval, opts []NeosyncTypeOption, arrayOpts ...NeosyncTypeOption) (*NeosyncArray, error) {
	neosyncAdapters := make([]NeosyncAdapter, len(elements))
	for i, e := range elements {
		neosyncAdapters[i] = NewInterval(opts...)
		err := neosyncAdapters[i].ScanPgx(e)
		if err != nil {
			return nil, err
		}
	}
	arrayAdapter := NewNeosyncArray(neosyncAdapters)
	return arrayAdapter, nil

}
