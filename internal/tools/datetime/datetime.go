package datetime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Unit is the timestamp resolution the user chose. It is never inferred
// from how many digits the number has.
type Unit int

const (
	Second Unit = iota
	Millisecond
)

const dateTimeLayout = "2006-01-02 15:04:05"

// ParseTimestamp reads a base-10 integer as seconds or milliseconds.
func ParseTimestamp(input string, unit Unit) (time.Time, error) {
	s := strings.TrimSpace(input)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的时间戳")
	}
	if unit == Millisecond {
		return time.UnixMilli(n), nil
	}
	return time.Unix(n, 0), nil
}

// ParseDate reads a date. A value without a zone is read in loc.
func ParseDate(input string, loc *time.Location) (time.Time, error) {
	s := strings.TrimSpace(input)
	if loc == nil {
		loc = time.Local
	}
	zoned := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range zoned {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	plain := []string{
		dateTimeLayout,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range plain {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无效的日期")
}

// FormatUnit renders t as a decimal timestamp in unit.
func FormatUnit(t time.Time, unit Unit) string {
	if unit == Millisecond {
		return strconv.FormatInt(t.UnixMilli(), 10)
	}
	return strconv.FormatInt(t.Unix(), 10)
}

// FormatLocal renders t in loc as a wall time without a numeric offset.
func FormatLocal(t time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	return t.In(loc).Format(dateTimeLayout)
}

// FormatZoned renders t in loc with its zone abbreviation.
func FormatZoned(t time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	return t.In(loc).Format("2006-01-02 15:04:05 MST")
}
