// Package gold reads the XAUS spot quote and daily closes, and turns them
// into the weight, currency, and lookback window the gold tool shows.
package gold

import (
	"time"
)

const (
	// GramsPerOunce is the troy ounce, which is what an "盎司" of gold is.
	GramsPerOunce = 31.1034768
	// SpotLimit is how old a quote can be and still count as current.
	SpotLimit = 5 * time.Minute
	// HistoryLimit is how old a daily series can be and still count as current.
	HistoryLimit = 24 * time.Hour
	// PollEvery is how often a visible quote is refreshed.
	PollEvery = 60 * time.Second
	// RequestTimeout bounds one request.
	RequestTimeout = 15 * time.Second
)

// Window is a lookback over the daily closes. Values match model.GoldWindow*.
type Window int

const (
	WindowMonth Window = iota
	WindowQuarter
	WindowHalf
	WindowYear
	WindowFiveYear
)

// Close is one completed daily close, in US dollars per troy ounce.
type Close struct {
	Date     time.Time
	USDPerOz float64
}

// Quote is one spot quote. Price is already in Currency and Unit.
// FX converts one US dollar into Currency. Stale means the feed itself
// marked the quote as its last real price.
type Quote struct {
	OK       bool
	Price    float64
	Currency string
	Unit     string
	USDPerOz float64
	FX       float64
	At       time.Time
	Stale    bool
}

// History is the daily series. Stale means the feed marked it so.
type History struct {
	Closes []Close
	At     time.Time
	Stale  bool
}

var currencies = []struct {
	code  string
	label string
	price string
}{
	{"CNY", "人民币", "元"},
	{"USD", "美元", "美元"},
	{"EUR", "欧元", "欧元"},
	{"GBP", "英镑", "英镑"},
	{"JPY", "日元", "日元"},
	{"HKD", "港币", "港币"},
}

// CurrencyLabels is the currency control, in order.
func CurrencyLabels() []string {
	out := make([]string, len(currencies))
	for i, item := range currencies {
		out[i] = item.label
	}
	return out
}

// CodeOf returns the ISO code for a currency label.
func CodeOf(label string) (string, bool) {
	for _, item := range currencies {
		if item.label == label {
			return item.code, true
		}
	}
	return "", false
}

// LabelOf returns the currency control label for an ISO code.
func LabelOf(code string) (string, bool) {
	for _, item := range currencies {
		if item.code == code {
			return item.label, true
		}
	}
	return "", false
}

// PriceName is the word used in a price, such as 元 for CNY.
func PriceName(code string) string {
	for _, item := range currencies {
		if item.code == code {
			return item.price
		}
	}
	return code
}

// WeightNames is the weight control, in order: 克, 盎司, 千克.
func WeightNames() []string {
	return []string{"克", "盎司", "千克"}
}

// WeightUnit is the API unit for a weight control index.
func WeightUnit(index int) string {
	switch index {
	case 1:
		return "oz"
	case 2:
		return "kg"
	default:
		return "gram"
	}
}

// WeightName is the word used in a price for an API unit.
func WeightName(unit string) string {
	switch unit {
	case "oz":
		return "盎司"
	case "kg":
		return "千克"
	default:
		return "克"
	}
}

// WindowNames is the lookback control, in order.
func WindowNames() []string {
	return []string{"1 个月", "3 个月", "6 个月", "1 年", "5 年"}
}

// WindowCode is the saved code for a lookback control index.
// An unknown index returns an empty string.
func WindowCode(index int) string {
	codes := []string{"1mo", "3mo", "6mo", "1y", "5y"}
	if index < 0 || index >= len(codes) {
		return ""
	}
	return codes[index]
}

// NormalizeWindow clamps index to a lookback. An unknown index is 1 年.
func NormalizeWindow(index int) Window {
	if index < 0 || index > int(WindowFiveYear) {
		return WindowYear
	}
	return Window(index)
}

// Convert turns a US dollar per troy ounce price into unit.
func Convert(usdPerOz float64, unit string) float64 {
	switch unit {
	case "oz":
		return usdPerOz
	case "kg":
		return usdPerOz / GramsPerOunce * 1000
	default:
		return usdPerOz / GramsPerOunce
	}
}

// Amount is the quote in currency and unit.
// A quote only converts its weight. Another currency has no rate here.
func (q Quote) Amount(currency, unit string) (float64, bool) {
	if !q.OK || q.Currency != currency || q.USDPerOz <= 0 {
		return 0, false
	}
	if q.Unit == unit && q.Price > 0 {
		return q.Price, true
	}
	if q.FX <= 0 {
		return 0, false
	}
	return Convert(q.USDPerOz, unit) * q.FX, true
}

// Expired reports that at is older than limit, or that the feed marked it stale.
// A zero time is not expired unless it was marked.
func Expired(at, now time.Time, marked bool, limit time.Duration) bool {
	if marked {
		return true
	}
	if at.IsZero() {
		return false
	}
	return now.Sub(at) > limit
}

// WindowStart is the first civil day of window, inclusive, in UTC.
// Five years reports all, and the returned day is the zero time.
func WindowStart(today time.Time, window Window) (time.Time, bool) {
	months := -12
	switch window {
	case WindowFiveYear:
		return time.Time{}, true
	case WindowMonth:
		months = -1
	case WindowQuarter:
		months = -3
	case WindowHalf:
		months = -6
	}
	return addMonths(civil(today), months), false
}

// InWindow keeps closes inside the lookback, including the first day and
// today, and drops a close dated after today.
func InWindow(points []Close, today time.Time, window Window) []Close {
	end := civil(today)
	start, all := WindowStart(today, window)
	out := make([]Close, 0, len(points))
	for _, point := range points {
		day := civil(point.Date)
		if day.After(end) {
			continue
		}
		if !all && day.Before(start) {
			continue
		}
		out = append(out, point)
	}
	return out
}

// Series is the lookback converted into unit, oldest first.
func Series(points []Close, today time.Time, window Window, unit string) (dates []time.Time, values []float64) {
	kept := InWindow(points, today, window)
	dates = make([]time.Time, len(kept))
	values = make([]float64, len(kept))
	for i, point := range kept {
		dates[i] = civil(point.Date)
		values[i] = Convert(point.USDPerOz, unit)
	}
	return dates, values
}

func civil(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// addMonths moves t by months, keeping the day or using the target
// month's last day when that day does not exist.
func addMonths(t time.Time, months int) time.Time {
	year, month, day := t.Date()
	loc := t.Location()
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc).AddDate(0, months, 0)
	last := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, loc).Day()
	if day > last {
		day = last
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, loc)
}
