package ui

import (
	"strings"
	"time"

	"devbox/internal/model"
	"devbox/internal/tools/datetime"

	myui "github.com/egoist/mygo/ui"
)

// ClockState is the timestamp tool's fields. Unit and Zone are what the
// segmented controls bind: seconds or milliseconds, local or UTC.
type ClockState struct {
	Unit      int
	Zone      int
	Timestamp string
	Date      string
	Result    string
}

func datetimeWorkspace(a *model.App, clock *ClockState, c *myui.Context) {
	if !c.WindowHidden() {
		c.After(time.Second)
	}
	now := c.Now()
	t := c.Theme()
	myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
		myui.Column(c).Padding(16, 24).Gap(12).Children(func() {
			myui.Text(c, "当前时间").FontSize(15).Bold()
			myui.Text(c, "当前本地 "+datetime.FormatZoned(now, time.Local)).TextColor(t.TextMuted).Selectable()
			myui.Text(c, "当前 UTC "+datetime.FormatZoned(now, time.UTC)).TextColor(t.TextMuted).Selectable()
			myui.Text(c, "当前秒 "+datetime.FormatUnit(now, datetime.Second)).TextColor(t.TextMuted).Selectable()
			myui.Text(c, "当前毫秒 "+datetime.FormatUnit(now, datetime.Millisecond)).TextColor(t.TextMuted).Selectable()

			myui.Row(c).Gap(16).AlignItems(myui.Center).Children(func() {
				myui.Segmented(c, &clock.Unit, "秒", "毫秒").Label("单位")
				myui.Segmented(c, &clock.Zone, "本地", "UTC").Label("时区")
			})
			myui.Text(c, "时间戳按所选单位解释，不根据位数判断。").FontSize(12).TextColor(t.TextMuted)

			myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
				myui.TextInput(c, &clock.Timestamp).Placeholder("Unix 时间戳").Label("时间戳").Width(280)
				if myui.Button(c, "转为日期").Clicked() {
					timestampToDate(a, clock)
				}
				if myui.Button(c, "使用当前时间").Clicked() {
					useNow(a, clock, now)
				}
			})
			myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
				myui.TextInput(c, &clock.Date).Placeholder("2006-01-02 15:04:05").Label("日期").Width(280)
				if myui.Button(c, "转为时间戳").Clicked() {
					dateToTimestamp(a, clock)
				}
				if myui.Button(c, "复制").Clicked() {
					copyClock(a, clock, c)
				}
			})
			if clock.Result != "" {
				myui.Text(c, "结果").FontSize(15).Bold()
				for _, line := range strings.Split(clock.Result, "\n") {
					myui.Text(c, line).Selectable()
				}
			}
		})
	})
}

func timestampToDate(a *model.App, clock *ClockState) {
	start := time.Now()
	tm, err := datetime.ParseTimestamp(clock.Timestamp, clockUnit(clock))
	finish(a, start, err, "已转为日期")
	if err != nil {
		return
	}
	clock.Date = datetime.FormatLocal(tm, clockLocation(clock))
	clock.Result = describeTime(tm)
}

func dateToTimestamp(a *model.App, clock *ClockState) {
	start := time.Now()
	tm, err := datetime.ParseDate(clock.Date, clockLocation(clock))
	finish(a, start, err, "已转为时间戳")
	if err != nil {
		return
	}
	clock.Timestamp = datetime.FormatUnit(tm, clockUnit(clock))
	clock.Result = describeTime(tm)
}

func useNow(a *model.App, clock *ClockState, now time.Time) {
	clock.Timestamp = datetime.FormatUnit(now, clockUnit(clock))
	clock.Date = datetime.FormatLocal(now, clockLocation(clock))
	clock.Result = describeTime(now)
	a.Status = "已填入当前时间"
	a.Error = ""
	a.Elapsed = ""
}

func copyClock(a *model.App, clock *ClockState, c *myui.Context) {
	if clock.Result == "" {
		a.Status = "没有结果"
		a.Error = ""
		a.Elapsed = ""
		return
	}
	c.WriteClipboard(clock.Result)
	a.Status = "已复制"
	a.Error = ""
	a.Elapsed = ""
}

func describeTime(tm time.Time) string {
	return strings.Join([]string{
		"本地 " + datetime.FormatZoned(tm, time.Local),
		"UTC " + datetime.FormatZoned(tm, time.UTC),
		"秒 " + datetime.FormatUnit(tm, datetime.Second),
		"毫秒 " + datetime.FormatUnit(tm, datetime.Millisecond),
	}, "\n")
}

func clockUnit(clock *ClockState) datetime.Unit {
	if clock.Unit == 1 {
		return datetime.Millisecond
	}
	return datetime.Second
}

func clockLocation(clock *ClockState) *time.Location {
	if clock.Zone == 1 {
		return time.UTC
	}
	return time.Local
}
