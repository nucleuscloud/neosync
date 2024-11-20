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

func (n NeosyncInterval) JsonScan(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &n)
	case string:
		return json.Unmarshal([]byte(v), &n)
	default:
		return fmt.Errorf("unsupported type for NeosyncInterval: %T", value)
	}
}

func (n NeosyncInterval) JsonValue() ([]byte, error) {
	return json.Marshal(n)
}
