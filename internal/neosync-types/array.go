package neosynctypes

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
)

type NeosyncArray struct {
	BaseType `                 json:",inline"`
	Elements []NeosyncAdapter `json:"elements"`
}

func NewNeosyncArray(
	elements []NeosyncAdapter,
	opts ...NeosyncTypeOption,
) (*NeosyncArray, error) {
	pgArray := &NeosyncArray{
		Elements: elements,
	}
	pgArray.Neosync.TypeId = NeosyncArrayId
	pgArray.setVersion(LatestVersion)

	if err := applyOptions(pgArray, opts...); err != nil {
		return nil, err
	}

	return pgArray, nil
}

func (a *NeosyncArray) setVersion(v Version) {
	a.Neosync.Version = v
}

func (a *NeosyncArray) GetVersion() Version {
	return a.Neosync.Version
}

func (a *NeosyncArray) ScanPgx(value any) error {
	valueSlice, err := gotypeutil.ParseSlice(value)
	if err != nil {
		return err
	}
	if len(valueSlice) != len(a.Elements) {
		return fmt.Errorf(
			"length mismatch: got %d elements, expected %d",
			len(valueSlice),
			len(a.Elements),
		)
	}
	for i, v := range valueSlice {
		if err := a.Elements[i].ScanPgx(v); err != nil {
			return fmt.Errorf("scanning element %d: %w", i, err)
		}
	}
	return nil
}

func (a *NeosyncArray) ValuePgx() (any, error) {
	values := make([]any, len(a.Elements))
	for i, e := range a.Elements {
		v, err := e.ValuePgx()
		if err != nil {
			return nil, fmt.Errorf("getting value for element %d: %w", i, err)
		}
		values[i] = v
	}
	return pq.Array(values), nil
}

func (a *NeosyncArray) ScanJson(value any) error {
	valueSlice, err := gotypeutil.ParseSlice(value)
	if err != nil {
		return err
	}
	if len(valueSlice) != len(a.Elements) {
		return fmt.Errorf(
			"length mismatch: got %d elements, expected %d",
			len(valueSlice),
			len(a.Elements),
		)
	}
	for i, v := range valueSlice {
		if err := a.Elements[i].ScanJson(v); err != nil {
			return fmt.Errorf("scanning element %d: %w", i, err)
		}
	}
	return nil
}

func (a *NeosyncArray) ValueJson() (any, error) {
	values := make([]any, len(a.Elements))
	for i, e := range a.Elements {
		v, err := e.ValueJson()
		if err != nil {
			return nil, fmt.Errorf("getting value for element %d: %w", i, err)
		}
		values[i] = v
	}
	return values, nil
}

func (a *NeosyncArray) ScanMysql(value any) error {
	valueSlice, err := gotypeutil.ParseSlice(value)
	if err != nil {
		return err
	}
	if len(valueSlice) != len(a.Elements) {
		return fmt.Errorf(
			"length mismatch: got %d elements, expected %d",
			len(valueSlice),
			len(a.Elements),
		)
	}
	for i, v := range valueSlice {
		if err := a.Elements[i].ScanMysql(v); err != nil {
			return fmt.Errorf("scanning element %d: %w", i, err)
		}
	}
	return nil
}

func (a *NeosyncArray) ValueMysql() (any, error) {
	values := make([]any, len(a.Elements))
	for i, e := range a.Elements {
		v, err := e.ValueMysql()
		if err != nil {
			return nil, fmt.Errorf("getting value for element %d: %w", i, err)
		}
		values[i] = v
	}
	return values, nil
}

func (a *NeosyncArray) ScanMssql(value any) error {
	return a.ScanMysql(value)
}

func (a *NeosyncArray) ValueMssql() (any, error) {
	return a.ValueMysql()
}
