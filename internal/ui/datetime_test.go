package ui

import (
	"strconv"
	"testing"
	"time"

	"devbox/internal/model"
)

func TestTimestampConvertsChosenUnit(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("时间戳"); err != nil {
		t.Fatal(err)
	}
	editors.Clock.Zone = 1
	editors.Clock.Unit = 1
	editors.Clock.Timestamp = "1000"
	tt.Frame()

	if err := tt.Click("转为日期"); err != nil {
		t.Fatal(err)
	}
	if editors.Clock.Date != "1970-01-01 00:00:01" || a.Status != "已转为日期" {
		t.Fatalf("date %q status %s result %q", editors.Clock.Date, a.Status, editors.Clock.Result)
	}
	if !tt.HasText("UTC 1970-01-01 00:00:01 UTC") || !tt.HasText("秒 1") || !tt.HasText("毫秒 1000") {
		t.Fatalf("texts %q", tt.Texts())
	}

	editors.Clock.Unit = 0
	editors.Clock.Timestamp = "1000"
	tt.Frame()
	if err := tt.Click("转为日期"); err != nil {
		t.Fatal(err)
	}
	if editors.Clock.Date != "1970-01-01 00:16:40" {
		t.Fatalf("seconds were read as milliseconds: %q", editors.Clock.Date)
	}
}

func TestDateConvertsInChosenZone(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("时间戳"); err != nil {
		t.Fatal(err)
	}
	editors.Clock.Zone = 1
	editors.Clock.Unit = 0
	editors.Clock.Date = "1970-01-01T08:00:00+08:00"
	tt.Frame()

	if err := tt.Click("转为时间戳"); err != nil {
		t.Fatal(err)
	}
	if editors.Clock.Timestamp != "0" || a.Status != "已转为时间戳" {
		t.Fatalf("timestamp %q status %s", editors.Clock.Timestamp, a.Status)
	}

	editors.Clock.Date = "not a date"
	tt.Frame()
	if err := tt.Click("转为时间戳"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "失败" || a.Error != "无效的日期" || editors.Clock.Timestamp != "0" {
		t.Fatalf("status %s error %s timestamp %q", a.Status, a.Error, editors.Clock.Timestamp)
	}
}

func TestCurrentTimeFillsAndTicks(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("时间戳"); err != nil {
		t.Fatal(err)
	}
	editors.Clock.Zone = 1
	editors.Clock.Unit = 0
	tt.Frame()

	before := tt.Now().Unix()
	if !tt.HasText("当前秒 " + strconv.FormatInt(before, 10)) {
		t.Fatalf("texts %q", tt.Texts())
	}
	if err := tt.Click("使用当前时间"); err != nil {
		t.Fatal(err)
	}
	if editors.Clock.Timestamp != strconv.FormatInt(before, 10) || a.Status != "已填入当前时间" {
		t.Fatalf("timestamp %q status %s now %d", editors.Clock.Timestamp, a.Status, before)
	}

	tt.Advance(2 * time.Second)
	if !tt.HasText("当前秒 "+strconv.FormatInt(tt.Now().Unix(), 10)) || tt.Now().Unix() == before {
		t.Fatalf("now %d texts %q", tt.Now().Unix(), tt.Texts())
	}
}
