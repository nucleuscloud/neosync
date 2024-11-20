package neosynctypes

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// scan from source to neosync type
// value from neosync type to destination

type NeosyncInterval struct {
	Microseconds int64 `json:"microseconds"`
	Days         int32 `json:"days"`
	Months       int32 `json:"months"`
}

func (n *NeosyncInterval) PgxScan(value *pgtype.Interval) error {
	if value == nil || !value.Valid {
		return nil
	}
	n.Microseconds = value.Microseconds
	n.Days = value.Days
	n.Months = value.Months
	return nil
}

func (n NeosyncInterval) PgxValue() (*pgtype.Interval, error) {
	return &pgtype.Interval{
		Microseconds: n.Microseconds,
		Days:         n.Days,
		Months:       n.Months,
	}, nil
}
