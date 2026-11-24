package model

const (
	GoldGram = iota
	GoldOunce
	GoldKilogram
)

const (
	GoldWindowMonth = iota
	GoldWindowQuarter
	GoldWindowHalf
	GoldWindowYear
	GoldWindowFiveYear
)

// GoldCurrencyCNY is the currency selected when none is saved.
const GoldCurrencyCNY = "人民币"

// goldCurrencies is the currency control. GoldCurrencyCNY is its first entry.
var goldCurrencies = []string{
	GoldCurrencyCNY,
	"美元",
	"欧元",
	"英镑",
	"日元",
	"港币",
}

// Gold is the gold tool's saved choices. Weight and Window are control indexes.
type Gold struct {
	Weight   int
	Currency string
	Window   int
}

// GoldCurrencies returns the currencies the gold page can select.
// The result is the shared list; callers must not modify it.
func GoldCurrencies() []string {
	return goldCurrencies
}

// ApplyDefaults replaces an unknown weight, currency, or window.
// A zero weight is 克, which is a real choice. A zero window is 1 个月,
// so the caller sets 1 年 before this when nothing was saved.
func (g *Gold) ApplyDefaults() {
	if g.Weight < GoldGram || g.Weight > GoldKilogram {
		g.Weight = GoldGram
	}
	if !knownGoldCurrency(g.Currency) {
		g.Currency = GoldCurrencyCNY
	}
	if g.Window < GoldWindowMonth || g.Window > GoldWindowFiveYear {
		g.Window = GoldWindowYear
	}
}

// UseGold stores saved codes. Empty or unknown codes become the defaults:
// 克, 人民币, and 1 年.
func (a *App) UseGold(weight, currency, window string) {
	a.Gold.Weight = goldWeightIndex(weight)
	a.Gold.Currency = goldCurrencyLabel(currency)
	a.Gold.Window = goldWindowIndex(window)
	a.Gold.ApplyDefaults()
}

// GoldWeightCode is the saved weight code.
func GoldWeightCode(index int) string {
	switch index {
	case GoldOunce:
		return "oz"
	case GoldKilogram:
		return "kg"
	default:
		return "gram"
	}
}

// GoldCurrencyCode is the saved ISO code. An unknown label returns an empty string.
func GoldCurrencyCode(label string) string {
	switch label {
	case "人民币":
		return "CNY"
	case "美元":
		return "USD"
	case "欧元":
		return "EUR"
	case "英镑":
		return "GBP"
	case "日元":
		return "JPY"
	case "港币":
		return "HKD"
	default:
		return ""
	}
}

// GoldWindowCode is the saved lookback code. An unknown index returns an empty string.
func GoldWindowCode(index int) string {
	switch index {
	case GoldWindowMonth:
		return "1mo"
	case GoldWindowQuarter:
		return "3mo"
	case GoldWindowHalf:
		return "6mo"
	case GoldWindowYear:
		return "1y"
	case GoldWindowFiveYear:
		return "5y"
	default:
		return ""
	}
}

func goldWeightIndex(code string) int {
	switch code {
	case "gram":
		return GoldGram
	case "oz":
		return GoldOunce
	case "kg":
		return GoldKilogram
	default:
		return -1
	}
}

func goldCurrencyLabel(code string) string {
	switch code {
	case "CNY":
		return "人民币"
	case "USD":
		return "美元"
	case "EUR":
		return "欧元"
	case "GBP":
		return "英镑"
	case "JPY":
		return "日元"
	case "HKD":
		return "港币"
	default:
		return ""
	}
}

func goldWindowIndex(code string) int {
	switch code {
	case "1mo":
		return GoldWindowMonth
	case "3mo":
		return GoldWindowQuarter
	case "6mo":
		return GoldWindowHalf
	case "1y":
		return GoldWindowYear
	case "5y":
		return GoldWindowFiveYear
	default:
		return -1
	}
}

func knownGoldCurrency(label string) bool {
	for _, item := range goldCurrencies {
		if item == label {
			return true
		}
	}
	return false
}
