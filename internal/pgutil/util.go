package pgutil

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDString(value pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", value.Bytes[0:4], value.Bytes[4:6], value.Bytes[6:8], value.Bytes[8:10], value.Bytes[10:16])
}

func UUIDStrings(values []pgtype.UUID) []string {
	outputs := []string{}
	for _, value := range values {
		outputs = append(outputs, UUIDString(value))
	}
	return outputs
}

func ToUuid(value string) (pgtype.UUID, error) {
	uuid := pgtype.UUID{}
	err := uuid.Scan(value)
	return uuid, err
}

func ToTimestamp(value time.Time) (pgtype.Timestamp, error) {
	timestamp := pgtype.Timestamp{}
	err := timestamp.Scan(value)
	return timestamp, err
}

func ToNullableString(text pgtype.Text) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}

func Int16ToBool(val int16) bool {
	return val > 0
}

func BoolToInt16(val bool) int16 {
	if val {
		return 1
	}
	return 0
}
