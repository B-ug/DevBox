package datetime

import (
	"testing"
	"time"
)

func TestParseTimestampUsesChosenUnit(t *testing.T) {
	sec, err := ParseTimestamp("1000", Second)
	if err != nil || sec.Unix() != 1000 {
		t.Fatalf("seconds %v %v", sec, err)
	}
	ms, err := ParseTimestamp("1000", Millisecond)
	if err != nil || ms.Unix() != 1 {
		t.Fatalf("milliseconds %v %v", ms, err)
	}
	if _, err := ParseTimestamp("12.5", Second); err == nil {
		t.Fatal("fractional timestamp was accepted")
	}
	if _, err := ParseTimestamp("  ", Second); err == nil {
		t.Fatal("blank timestamp was accepted")
	}
}

func TestParseDate(t *testing.T) {
	utc, err := ParseDate("1970-01-01 00:00:01", time.UTC)
	if err != nil || utc.Unix() != 1 {
		t.Fatalf("utc %v %v", utc, err)
	}
	zoned, err := ParseDate("1970-01-01T08:00:00+08:00", time.UTC)
	if err != nil || zoned.Unix() != 0 {
		t.Fatalf("zoned %v %v", zoned, err)
	}
	day, err := ParseDate("1970-01-02", time.UTC)
	if err != nil || day.Unix() != 86400 {
		t.Fatalf("day %v %v", day, err)
	}
	if _, err := ParseDate("yesterday", time.UTC); err == nil {
		t.Fatal("invalid date was accepted")
	}
}

func TestFormat(t *testing.T) {
	tm := time.Unix(0, 0)
	if FormatUnit(tm, Second) != "0" || FormatUnit(tm, Millisecond) != "0" {
		t.Fatalf("units %s %s", FormatUnit(tm, Second), FormatUnit(tm, Millisecond))
	}
	if FormatZoned(tm, time.UTC) != "1970-01-01 00:00:00 UTC" {
		t.Fatalf("zoned %s", FormatZoned(tm, time.UTC))
	}
	if FormatLocal(tm, time.UTC) != "1970-01-01 00:00:00" {
		t.Fatalf("local %s", FormatLocal(tm, time.UTC))
	}
}
