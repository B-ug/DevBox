package gold

import (
	"strconv"
	"strings"
	"time"
)

// FormatAmount renders amount with thousands separators.
// Grams keep four decimal places. Ounces and kilograms keep two.
func FormatAmount(amount float64, unit string) string {
	decimals := 2
	if unit != "oz" && unit != "kg" {
		decimals = 4
	}
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	raw := strconv.FormatFloat(amount, 'f', decimals, 64)
	whole, frac, _ := strings.Cut(raw, ".")
	return sign + groupDigits(whole) + "." + frac
}

// FormatQuote renders a price as "788.9496 元/克".
func FormatQuote(amount float64, currency, unit string) string {
	return FormatAmount(amount, unit) + " " + PriceName(currency) + "/" + WeightName(unit)
}

// FormatDay renders a civil day as "2026年11月23日".
func FormatDay(day time.Time) string {
	year, month, date := day.Date()
	return strconv.Itoa(year) + "年" + strconv.Itoa(int(month)) + "月" + strconv.Itoa(date) + "日"
}

// FormatQuoteTime renders the quote instant in loc as "2026年11月24日 09:41".
func FormatQuoteTime(at time.Time, loc *time.Location) string {
	if at.IsZero() {
		return ""
	}
	if loc == nil {
		loc = time.UTC
	}
	at = at.In(loc)
	return FormatDay(at) + " " + two(at.Hour()) + ":" + two(at.Minute())
}

// FormatReadout renders one daily close as "2026年11月23日 117.1800 美元/克".
// amount is already in unit.
func FormatReadout(day time.Time, amount float64, unit string) string {
	return FormatDay(day) + " " + FormatAmount(amount, unit) + " 美元/" + WeightName(unit)
}

func groupDigits(whole string) string {
	if whole == "" {
		return "0"
	}
	n := len(whole)
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(whole[i])
	}
	return b.String()
}

func two(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
