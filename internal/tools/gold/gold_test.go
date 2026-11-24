package gold

import (
	"math"
	"testing"
	"time"
)

func TestWindowStartClampsMonthEnd(t *testing.T) {
	cases := []struct {
		today  time.Time
		window Window
		want   string
		all    bool
	}{
		{time.Date(2023, 3, 31, 15, 0, 0, 0, time.FixedZone("CST", 8*3600)), WindowMonth, "2023-02-28", false},
		{time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC), WindowMonth, "2024-02-29", false},
		{time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), WindowYear, "2023-02-28", false},
		{time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC), WindowMonth, "2025-12-31", false},
		{time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), WindowHalf, "2026-02-28", false},
		{time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC), WindowQuarter, "2026-08-24", false},
		{time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC), WindowFiveYear, "", true},
	}
	for _, tc := range cases {
		start, all := WindowStart(tc.today, tc.window)
		if all != tc.all {
			t.Fatalf("%s all %v", tc.today, all)
		}
		if tc.all {
			continue
		}
		if got := start.Format("2006-01-02"); got != tc.want || start.Location() != time.UTC {
			t.Fatalf("%s window %d got %s", tc.today.Format("2006-01-02"), tc.window, got)
		}
	}
}

func TestInWindowUsesCalendarDays(t *testing.T) {
	today := time.Date(2026, 11, 24, 22, 0, 0, 0, time.FixedZone("CST", 8*3600))
	points := []Close{
		{Date: day(2021, 11, 23), USDPerOz: 1},
		{Date: day(2025, 11, 23), USDPerOz: 2},
		{Date: day(2025, 11, 24), USDPerOz: 3},
		{Date: day(2026, 10, 23), USDPerOz: 4},
		{Date: day(2026, 10, 24), USDPerOz: 5},
		{Date: day(2026, 11, 23), USDPerOz: 6},
		{Date: day(2026, 11, 24), USDPerOz: 7},
		{Date: day(2026, 11, 25), USDPerOz: 8},
	}
	got := prices(InWindow(points, today, WindowMonth))
	if !same(got, []float64{5, 6, 7}) {
		t.Fatalf("month %v", got)
	}
	got = prices(InWindow(points, today, WindowYear))
	if !same(got, []float64{3, 4, 5, 6, 7}) {
		t.Fatalf("year %v", got)
	}
	got = prices(InWindow(points, today, WindowFiveYear))
	if !same(got, []float64{1, 2, 3, 4, 5, 6, 7}) {
		t.Fatalf("five %v", got)
	}
}

func TestConvertAndAmount(t *testing.T) {
	const usd = 3110.34768
	if math.Abs(Convert(usd, "gram")-100) > 1e-9 || math.Abs(Convert(usd, "kg")-100000) > 1e-6 || Convert(usd, "oz") != usd {
		t.Fatalf("gram %v kg %v oz %v", Convert(usd, "gram"), Convert(usd, "kg"), Convert(usd, "oz"))
	}
	quote := Quote{OK: true, Price: 700, Currency: "CNY", Unit: "gram", USDPerOz: usd, FX: 7}
	if amount, ok := quote.Amount("CNY", "gram"); !ok || amount != 700 {
		t.Fatalf("same unit %v %v", amount, ok)
	}
	amount, ok := quote.Amount("CNY", "oz")
	if !ok || math.Abs(amount-usd*7) > 1e-6 {
		t.Fatalf("converted %v %v", amount, ok)
	}
	if _, ok := quote.Amount("USD", "oz"); ok {
		t.Fatal("converted currency")
	}
}

func TestFormat(t *testing.T) {
	if got := FormatQuote(788.9496, "CNY", "gram"); got != "788.9496 元/克" {
		t.Fatalf("quote %s", got)
	}
	if got := FormatAmount(3673.9, "oz"); got != "3,673.90" {
		t.Fatalf("oz %s", got)
	}
	if got := FormatAmount(3110347.68, "kg"); got != "3,110,347.68" {
		t.Fatalf("kg %s", got)
	}
	day := day(2026, 11, 23)
	if got := FormatReadout(day, 117.18, "gram"); got != "2026年11月23日 117.1800 美元/克" {
		t.Fatalf("readout %s", got)
	}
	at := time.Date(2026, 11, 24, 1, 41, 0, 0, time.UTC)
	loc := time.FixedZone("CST", 8*3600)
	if got := FormatQuoteTime(at, loc); got != "2026年11月24日 09:41" {
		t.Fatalf("time %s", got)
	}
}

func TestExpired(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	if Expired(now.Add(-5*time.Minute), now, false, SpotLimit) {
		t.Fatal("five minutes is still current")
	}
	if !Expired(now.Add(-5*time.Minute-time.Second), now, false, SpotLimit) {
		t.Fatal("older than five minutes")
	}
	if !Expired(now, now, true, SpotLimit) {
		t.Fatal("marked stale")
	}
	if Expired(time.Time{}, now, false, SpotLimit) {
		t.Fatal("missing time")
	}
	if Expired(now.Add(-24*time.Hour), now, false, HistoryLimit) || !Expired(now.Add(-24*time.Hour-time.Second), now, false, HistoryLimit) {
		t.Fatal("history limit")
	}
}

func day(year int, month time.Month, date int) time.Time {
	return time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
}

func prices(points []Close) []float64 {
	out := make([]float64, len(points))
	for i, point := range points {
		out[i] = point.USDPerOz
	}
	return out
}

func same(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
