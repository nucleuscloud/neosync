package neosynctypes

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Bit struct {
	BaseType    `json:",inline"`
	JsonScanner `json:"-"`
	Bytes       []byte `json:"bytes"`
	Len         int32  `json:"len"`
	// Bit   string `json:"bit_string"`
}

func (b *Bit) ScanPgx(value any) error {
	if value == nil {
		return nil
	}
	Bit, ok := value.(*pgtype.Bits)
	if !ok {
		return fmt.Errorf("expected *pgtype.Bit, got %T", value)
	}
	b.Bytes = Bit.Bytes
	b.Len = Bit.Len
	return nil
}

func (b *Bit) ValuePgx() (any, error) {
	return &pgtype.Bits{
		Bytes: b.Bytes,
		Len:   b.Len,
	}, nil
}

func (b *Bit) ScanMysql(value any) error {
	return errors.ErrUnsupported
}

func (b *Bit) ValueMysql() (any, error) {
	return nil, errors.ErrUnsupported
}

func (b *Bit) ScanJson(value any) error {
	return b.JsonScanner.ScanJson(value, b)
}

func (b *Bit) ValueJson() (any, error) {
	return b.JsonScanner.ValueJson(b)
}

func (b *Bit) setVersion(v Version) {
	b.Neosync.Version = v
}

func (b *Bit) GetVersion() Version {
	return b.Neosync.Version
}

func NewBitFromPgx(value any, opts ...NeosyncTypeOption) (*Bit, error) {
	Bit, err := NewBit(opts...)
	if err != nil {
		return nil, err
	}
	err = Bit.ScanPgx(value)
	if err != nil {
		return nil, err
	}
	return Bit, nil
}

func NewBit(opts ...NeosyncTypeOption) (*Bit, error) {
	bit := &Bit{}
	bit.Neosync.TypeId = NeosyncBitId
	bit.setVersion(LatestVersion)

	if err := applyOptions(bit, opts...); err != nil {
		return nil, err
	}
	return bit, nil
}

func NewBitArrayFromPgx(elements []*pgtype.Bits, opts []NeosyncTypeOption, arrayOpts ...NeosyncTypeOption) (*NeosyncArray, error) {
	neosyncAdapters := make([]NeosyncAdapter, len(elements))
	for i, e := range elements {
		newBit, err := NewBit(opts...)
		if err != nil {
			return nil, err
		}
		neosyncAdapters[i] = newBit
		err = neosyncAdapters[i].ScanPgx(e)
		if err != nil {
			return nil, err
		}
	}
	return NewNeosyncArray(neosyncAdapters)
}
