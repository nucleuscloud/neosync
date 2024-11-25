package neosynctypes

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Interval struct {
	BaseType     `json:",inline"`
	JsonScanner  `json:"-"`
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
	return i.JsonScanner.ScanJson(value, i)
}

func (i *Interval) ValueJson() (any, error) {
	return i.JsonScanner.ValueJson(i)
}

func (i *Interval) setVersion(v Version) {
	i.Neosync.Version = v
}

func (i *Interval) GetVersion() Version {
	return i.Neosync.Version
}

func NewIntervalFromPgx(value any, opts ...NeosyncTypeOption) (*Interval, error) {
	interval, err := NewInterval(opts...)
	if err != nil {
		return nil, err
	}
	err = interval.ScanPgx(value)
	if err != nil {
		return nil, err
	}
	return interval, nil
}

func NewInterval(opts ...NeosyncTypeOption) (*Interval, error) {
	interval := &Interval{}
	interval.Neosync.TypeId = NeosyncIntervalId
	interval.setVersion(LatestVersion)

	if err := applyOptions(interval, opts...); err != nil {
		return nil, err
	}
	return interval, nil
}

func NewIntervalArrayFromPgx(elements []*pgtype.Interval, opts []NeosyncTypeOption, arrayOpts ...NeosyncTypeOption) (*NeosyncArray, error) {
	neosyncAdapters := make([]NeosyncAdapter, len(elements))
	for i, e := range elements {
		newInterval, err := NewInterval(opts...)
		if err != nil {
			return nil, err
		}
		neosyncAdapters[i] = newInterval
		err = neosyncAdapters[i].ScanPgx(e)
		if err != nil {
			return nil, err
		}
	}
	return NewNeosyncArray(neosyncAdapters)

}
