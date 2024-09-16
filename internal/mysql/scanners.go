package mysql

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type MyDate struct {
	time.Time
}

func (d *MyDate) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("cannot scan type %T into MyDate", value)
	}

	if s == "0000-00-00" || s == "0000-00-00 00:00:00" {
		d.Time = time.Time{}
		return nil
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"02/01/2006",
		"02/01/2006 15:04:05",
		time.RFC3339,
	}

	var err error
	for _, layout := range layouts {
		d.Time, err = time.Parse(layout, s)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to parse date: %s", s)
}

func (d MyDate) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time.Format("2006-01-02 15:04:05"), nil
}
