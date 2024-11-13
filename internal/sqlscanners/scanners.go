package sqlscanners

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

type SqlScanner interface {
	Scan(value any) error
	Value() (driver.Value, error)
	String() string
}

type BitString struct {
	IsValid   bool
	Bytes     []byte
	BitString string
}

// Scan implements the sql.Scanner interface for BitString
func (b *BitString) Scan(value any) error {
	if value == nil {
		b.IsValid = false
		return nil
	}

	b.IsValid = true

	switch v := value.(type) {
	case []byte:
		b.Bytes = make([]byte, len(v))
		copy(b.Bytes, v)
		if len(v) == 1 {
			// BIT(1) to BIT(8)
			b.BitString = strconv.FormatUint(uint64(v[0]), 2)
		} else {
			// BIT(9) to BIT(64)
			val := uint64(0)
			for i, byte := range v {
				// Calculate the shift amount
				shiftAmount := 8 * (len(v) - 1 - i)
				// Convert byte to uint64 and shift it
				shiftedByte := uint64(byte) << shiftAmount
				// Bitwise OR
				val |= shiftedByte
			}
			b.BitString = strconv.FormatUint(val, 2)
		}
	case int64:
		b.BitString = strconv.FormatInt(v, 2)
		b.Bytes = []byte{byte(v)}
	case uint64:
		b.BitString = strconv.FormatUint(v, 2)
		b.Bytes = []byte{byte(v)}
	case string:
		b.BitString = v
		val, err := strconv.ParseUint(v, 2, 64)
		if err != nil {
			return fmt.Errorf("invalid binary string: %v", err)
		}
		b.Bytes = make([]byte, (len(v)+7)/8)
		for i := range b.Bytes {
			bitPosition := len(b.Bytes) - i - 1
			if bitPosition < 0 || bitPosition > len(b.Bytes) {
				return fmt.Errorf("error scanning bit. invalid bit position")
			}
			shiftAmount := uint(8) * uint(bitPosition)
			b.Bytes[i] = byte(val >> shiftAmount)
		}
	default:
		return fmt.Errorf("cannot scan type %T into BitString", value)
	}
	return nil
}

// Value implements the driver.Valuer interface for BitString
func (b BitString) Value() (driver.Value, error) {
	if !b.IsValid {
		return nil, nil
	}
	return b.Bytes, nil
}

func (b BitString) String() string {
	if !b.IsValid {
		return ""
	}
	return b.BitString
}
