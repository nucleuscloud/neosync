package neosynctypes

import (
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"
)

type Bits struct {
	BaseType    `json:",inline"`
	JsonScanner `json:"-"`
	Bytes       []byte `json:"bytes"`
	Len         int32  `json:"len"`
}

func (b *Bits) ScanPgx(value any) error {
	if value == nil {
		return nil
	}
	bits, ok := value.(*pgtype.Bits)
	if !ok {
		return fmt.Errorf("expected *pgtype.Bit, got %T", value)
	}
	b.Bytes = bits.Bytes
	b.Len = bits.Len
	return nil
}

func (b *Bits) ValuePgx() (any, error) {
	return &pgtype.Bits{
		Bytes: b.Bytes,
		Len:   b.Len,
		Valid: true,
	}, nil
}

func (b *Bits) ScanMysql(value any) error {
	if value == nil {
		return nil
	}
	bits, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected *sqlscanners.BitString, got %T", value)
	}
	if len(bits) > math.MaxInt32 {
		return fmt.Errorf("bit string length %d exceeds maximum int32 value", len(bits))
	}
	b.Bytes = bits
	b.Len = int32(len(bits)) //nolint:gosec
	return nil
}

func (b *Bits) ScanMssql(value any) error {
	return b.ScanMysql(value)
}

func (b *Bits) ValueMssql() (any, error) {
	return b.ValueMysql()
}

func (b *Bits) ValueMysql() (any, error) {
	return b.Bytes, nil
}

func (b *Bits) ScanJson(value any) error {
	return b.JsonScanner.ScanJson(value, b)
}

func (b *Bits) ValueJson() (any, error) {
	return b.JsonScanner.ValueJson(b)
}

func (b *Bits) setVersion(v Version) {
	b.Neosync.Version = v
}

func (b *Bits) GetVersion() Version {
	return b.Neosync.Version
}

func NewBitsFromPgx(value any, opts ...NeosyncTypeOption) (*Bits, error) {
	bits, err := NewBits(opts...)
	if err != nil {
		return nil, err
	}
	err = bits.ScanPgx(value)
	if err != nil {
		return nil, err
	}
	return bits, nil
}

func NewBitsFromMysql(value any, opts ...NeosyncTypeOption) (*Bits, error) {
	bits, err := NewBits(opts...)
	if err != nil {
		return nil, err
	}
	err = bits.ScanMysql(value)
	if err != nil {
		return nil, err
	}
	return bits, nil
}

func NewBitsFromMssql(value any, opts ...NeosyncTypeOption) (*Bits, error) {
	bits, err := NewBits(opts...)
	if err != nil {
		return nil, err
	}
	err = bits.ScanMssql(value)
	if err != nil {
		return nil, err
	}
	return bits, nil
}

func NewBits(opts ...NeosyncTypeOption) (*Bits, error) {
	bits := &Bits{}
	bits.Neosync.TypeId = NeosyncBitsId
	bits.setVersion(LatestVersion)

	if err := applyOptions(bits, opts...); err != nil {
		return nil, err
	}
	return bits, nil
}

func NewBitsArrayFromPgx(elements []*pgtype.Bits, opts []NeosyncTypeOption, arrayOpts ...NeosyncTypeOption) (*NeosyncArray, error) {
	neosyncAdapters := make([]NeosyncAdapter, len(elements))
	for i, e := range elements {
		newBits, err := NewBits(opts...)
		if err != nil {
			return nil, err
		}
		neosyncAdapters[i] = newBits
		err = neosyncAdapters[i].ScanPgx(e)
		if err != nil {
			return nil, err
		}
	}
	return NewNeosyncArray(neosyncAdapters)
}
