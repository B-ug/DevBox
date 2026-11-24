package ui

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"devbox/internal/config"
	"devbox/internal/model"
	"devbox/internal/tools/gold"

	myui "github.com/egoist/mygo/ui"
)

func TestGoldCatalogMatches(t *testing.T) {
	if !slices.Equal(model.GoldCurrencies(), gold.CurrencyLabels()) {
		t.Fatalf("model %v gold %v", model.GoldCurrencies(), gold.CurrencyLabels())
	}
	for i, label := range model.GoldCurrencies() {
		code := model.GoldCurrencyCode(label)
		back, ok := gold.LabelOf(code)
		got, known := gold.CodeOf(label)
		if !ok || !known || got != code || back != label {
			t.Fatalf("%d %s code %s back %s known %v labelok %v", i, label, got, back, known, ok)
		}
	}
	if !slices.Equal(gold.WeightNames(), []string{"克", "盎司", "千克"}) {
		t.Fatalf("weights %v", gold.WeightNames())
	}
	for i, unit := range []string{"gram", "oz", "kg"} {
		if model.GoldWeightCode(i) != unit || gold.WeightUnit(i) != unit {
			t.Fatalf("weight %d model %s gold %s", i, model.GoldWeightCode(i), gold.WeightUnit(i))
		}
	}
	if !slices.Equal(gold.WindowNames(), []string{"1 个月", "3 个月", "6 个月", "1 年", "5 年"}) {
		t.Fatalf("windows %v", gold.WindowNames())
	}
	for i, code := range []string{"1mo", "3mo", "6mo", "1y", "5y"} {
		if model.GoldWindowCode(i) != code || gold.WindowCode(i) != code {
			t.Fatalf("window %d model %s gold %s", i, model.GoldWindowCode(i), gold.WindowCode(i))
		}
	}
}

func TestGoldPageShowsQuoteAndCloses(t *testing.T) {
	var spot, hist atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/spot":
			spot.Add(1)
			unit := r.URL.Query().Get("unit")
			price := "788.9496"
			apiUnit := "gram"
			if unit == "oz" {
				price = "3673.90"
				apiUnit = "oz"
			}
			_, _ = w.Write([]byte(`{
				"xau": {"price": ` + price + `, "currency": "CNY", "unit": "` + apiUnit + `"},
				"spot_usd_oz": 3673.9,
				"fx_rate": 6.6793,
				"data_state": {"status": "fresh", "as_of": "` + time.Now().UTC().Format(time.RFC3339) + `"}
			}`))
		case "/api/v1/history":
			hist.Add(1)
			_, _ = w.Write([]byte(`{
				"points": [{"d": "2020-01-02", "c": 1500}, {"d": "` + time.Now().AddDate(0, 0, -1).Format("2006-01-02") + `", "c": 3600}],
				"data_state": {"status": "fresh", "as_of": "` + time.Now().UTC().Format(time.RFC3339) + `"}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	restoreGoldBase(t, server.URL)

	a := model.New()
	a.ConfigDir = t.TempDir()
	_, tt := layoutTester(a, 1200, 900)
	if err := tt.Click("黄金"); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"查看现货金价和历史走势",
		"788.9496 元/克",
		"重量",
		"币种",
		"人民币",
		"回看窗口",
		"1 年",
		"美元/克",
		"走势",
		"参考价，不能用于成交。来源 xaus.com",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if tt.HasText("正在获取金价") || tt.HasText("暂时拿不到金价") || tt.HasText("价格已过期") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if spot.Load() != 1 || hist.Load() != 1 {
		t.Fatalf("spot %d hist %d", spot.Load(), hist.Load())
	}

	if err := tt.Click("盎司"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("3,673.90 元/盎司") || a.Gold.Weight != model.GoldOunce {
		t.Fatalf("weight %d texts %q", a.Gold.Weight, tt.Texts())
	}
	file, err := config.Load(a.ConfigDir)
	if err != nil || file.Gold.Weight != "oz" || file.Gold.Currency != "CNY" || file.Gold.Window != "1y" {
		t.Fatalf("%+v err %v", file.Gold, err)
	}

	if err := tt.Click("5 年"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Gold.Window != model.GoldWindowFiveYear || hist.Load() != 1 {
		t.Fatalf("window %d hist %d", a.Gold.Window, hist.Load())
	}

	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(gold.PollEvery)
	if spot.Load() != 2 || hist.Load() != 1 {
		t.Fatalf("after leave spot %d hist %d", spot.Load(), hist.Load())
	}
}

func TestGoldPollsQuoteOnly(t *testing.T) {
	var spot, hist atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/history" {
			hist.Add(1)
			_, _ = w.Write([]byte(`{"points": [{"d": "2026-11-01", "c": 3600}], "data_state": {"status": "fresh", "as_of": "` + time.Now().UTC().Format(time.RFC3339) + `"}}`))
			return
		}
		spot.Add(1)
		_, _ = w.Write([]byte(`{
			"xau": {"price": 100, "currency": "CNY", "unit": "gram"},
			"spot_usd_oz": 3110.34768,
			"fx_rate": 7,
			"data_state": {"status": "fresh", "as_of": "` + time.Now().UTC().Format(time.RFC3339) + `"}
		}`))
	}))
	defer server.Close()
	restoreGoldBase(t, server.URL)

	a := model.New()
	_, tt := layoutTester(a, 1200, 900)
	if err := tt.Click("黄金"); err != nil {
		t.Fatal(err)
	}
	if spot.Load() != 1 || hist.Load() != 1 || !tt.HasText("100.0000 元/克") {
		t.Fatalf("spot %d hist %d texts %q", spot.Load(), hist.Load(), tt.Texts())
	}
	tt.Advance(gold.PollEvery)
	if spot.Load() != 2 || hist.Load() != 1 {
		t.Fatalf("spot %d hist %d", spot.Load(), hist.Load())
	}
}

func TestGoldFailureAndStale(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/spot" {
			_, _ = w.Write([]byte(`{
				"xau": {"price": 100, "currency": "CNY", "unit": "gram"},
				"spot_usd_oz": 3110.34768,
				"fx_rate": 7,
				"stale": true,
				"data_state": {"status": "stale", "as_of": "2020-01-01T00:00:00Z"}
			}`))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	restoreGoldBase(t, server.URL)

	a := model.New()
	_, tt := layoutTester(a, 1200, 900)
	if err := tt.Click("黄金"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("100.0000 元/克") || !tt.HasText("价格已过期") || !tt.HasText("暂时拿不到走势") {
		t.Fatalf("texts %q", tt.Texts())
	}

	s := newGoldState()
	s.onScreen = true
	s.spotGen = 2
	s.finishSpot(gold.Quote{OK: true, Price: 1, Currency: "CNY", Unit: "gram", USDPerOz: 1, FX: 1}, nil, 1)
	if s.quoteOK {
		t.Fatal("applied a stale quote")
	}
	s.spotGen = 3
	s.onScreen = false
	s.finishSpot(gold.Quote{OK: true, Price: 1, Currency: "CNY", Unit: "gram", USDPerOz: 1, FX: 1}, nil, 3)
	if s.quoteOK {
		t.Fatal("applied a quote after leaving")
	}
}

func TestGoldFirstFailureIsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	restoreGoldBase(t, server.URL)

	a := model.New()
	_, tt := layoutTester(a, 1200, 900)
	if err := tt.Click("黄金"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("暂时拿不到金价") || !tt.HasText("暂时拿不到走势") || tt.HasText("更新失败") {
		t.Fatalf("texts %q", tt.Texts())
	}
}

func TestGoldReadoutStaysInside(t *testing.T) {
	var start, end time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/history" {
			_, _ = w.Write([]byte(`{"points":[
				{"d":"` + start.Format("2006-01-02") + `","c":1000},
				{"d":"` + end.Format("2006-01-02") + `","c":2000}
			],"data_state":{"status":"fresh","as_of":"` + time.Now().UTC().Format(time.RFC3339) + `"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"xau":{"price":100,"currency":"CNY","unit":"gram"},"spot_usd_oz":3110.34768,"fx_rate":7,"data_state":{"status":"fresh","as_of":"` + time.Now().UTC().Format(time.RFC3339) + `"}}`))
	}))
	defer server.Close()
	restoreGoldBase(t, server.URL)

	a := model.New()
	_, tt := layoutTester(a, 1000, 800)
	now := tt.Now()
	start = now.AddDate(0, -2, 0)
	end = now.AddDate(0, 0, -1)
	if err := tt.Click("黄金"); err != nil {
		t.Fatal(err)
	}
	area, ok := tt.Find("走势图")
	if !ok {
		t.Fatalf("missing plot in %q", tt.Texts())
	}
	left := gold.FormatReadout(time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC), gold.Convert(1000, "gram"), "gram")
	right := gold.FormatReadout(time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC), gold.Convert(2000, "gram"), "gram")

	tt.Move(area.X+area.W-2, area.Y+area.H/2)
	if box, found := tt.Find(right); !found || !boxInside(box, area) {
		t.Fatalf("right %q box %+v found %v area %+v texts %q", right, box, found, area, tt.Texts())
	}
	tt.Move(area.X+2, area.Y+area.H/2)
	if box, found := tt.Find(left); !found || !boxInside(box, area) {
		t.Fatalf("left %q box %+v found %v area %+v", left, box, found, area)
	}
}

func boxInside(box, area myui.Rect) bool {
	const slop float32 = 2
	return box.X >= area.X-slop && box.Y >= area.Y-slop &&
		box.X+box.W <= area.X+area.W+slop && box.Y+box.H <= area.Y+area.H+slop
}

func restoreGoldBase(t *testing.T, base string) {
	t.Helper()
	previous := gold.BaseURL
	gold.BaseURL = base
	t.Cleanup(func() { gold.BaseURL = previous })
}
