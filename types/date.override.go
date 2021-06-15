package types

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

func (date *Date) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	time := nullTime.Time
	*date = Date{
		Year:  int32(time.Year()),
		Month: int32(time.Month()),
		Day:   int32(time.Day()),
	}
	return
}

func (date Date) Value() (driver.Value, error) {
	y, m, d := int(date.Year), time.Month(date.Month), int(date.Day)
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC), nil
}

// GormDataType gorm common data type
func (date Date) GormDataType() string {
	return "date"
}

func (date Date) GobEncode() ([]byte, error) {
	y, m, d := int(date.Year), time.Month(date.Month), int(date.Day)
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).GobEncode()
}

func (date *Date) GobDecode(b []byte) error {
	y, m, d := int(date.Year), time.Month(date.Month), int(date.Day)
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return (&t).GobDecode(b)
}

func (date Date) MarshalJSON() ([]byte, error) {
	y, m, d := int(date.Year), time.Month(date.Month), int(date.Day)
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return t.MarshalJSON()
}

func (date *Date) UnmarshalJSON(b []byte) error {
	y, m, d := int(date.Year), time.Month(date.Month), int(date.Day)
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return (&t).UnmarshalJSON(b)
}
