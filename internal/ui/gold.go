package ui

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"devbox/internal/model"
	"devbox/internal/tools/gold"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/chart"

	myui "github.com/egoist/mygo/ui"
)

// GoldState is the gold tool's live quote, daily closes, and the choices
// the controls bind. Choices are copied from the model on the first frame.
type GoldState struct {
	weight   int
	currency string
	window   int
	bound    bool
	selected int

	onScreen bool
	wantSpot bool
	wantHist bool

	spotGen     int
	spotCancel  context.CancelFunc
	spotRunning bool
	spotAt      time.Time
	quote       gold.Quote
	quoteOK     bool
	spotFailed  bool

	histGen     int
	histCancel  context.CancelFunc
	histRunning bool
	closes      []gold.Close
	histAt      time.Time
	histStale   bool
	histOK      bool
	histFailed  bool
}

func newGoldState() *GoldState {
	return &GoldState{selected: -1}
}

// Leave stops refresh. A result that arrives afterwards is ignored.
func (s *GoldState) Leave() {
	if s == nil || !s.onScreen {
		return
	}
	s.onScreen = false
	s.stopSpot()
	s.stopHist()
}

func goldWorkspace(a *model.App, editors *Editors, c *myui.Context) {
	s := editors.Gold
	if s == nil {
		s = newGoldState()
		editors.Gold = s
	}
	prepareGold(a, s, editors.window, c)
	t := c.Theme()
	now := c.Now()
	myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
		myui.Column(c).Padding(16, 24).Gap(12).MinWidth(0).Children(func() {
			goldChoices(a, s, editors.window, c, now)
			goldQuote(s, now, t, c)
			goldWindow(a, s, c)
			goldTrend(s, now, t, c)
			myui.Text(c, "参考价，不能用于成交。来源 xaus.com").FontSize(12).TextColor(t.TextMuted)
		})
	})
}

// prepareGold starts a fetch when the tool has just been shown, and refreshes
// the quote on a one minute cadence. The flags make the work run once for a
// visit rather than once per rebuild.
func prepareGold(a *model.App, s *GoldState, win *mygo.Window, c *myui.Context) {
	if !s.bound {
		a.Gold.ApplyDefaults()
		s.weight = a.Gold.Weight
		s.currency = a.Gold.Currency
		s.window = a.Gold.Window
		s.bound = true
		s.selected = -1
	}
	if !s.onScreen {
		s.onScreen = true
		s.wantSpot = true
		s.wantHist = true
		s.selected = -1
	}
	now := c.Now()
	if s.wantSpot && !s.spotRunning {
		s.wantSpot = false
		s.startSpot(win, now)
	}
	if s.wantHist && !s.histRunning {
		s.wantHist = false
		s.startHist(win)
	}
	if !s.spotRunning && !s.spotAt.IsZero() && !now.Before(s.spotAt.Add(gold.PollEvery)) {
		s.startSpot(win, now)
	}
	if c.WindowHidden() {
		return
	}
	wait := gold.PollEvery
	if !s.spotAt.IsZero() {
		if remaining := gold.PollEvery - now.Sub(s.spotAt); remaining < wait {
			wait = remaining
		}
	}
	if wait < time.Second {
		wait = time.Second
	}
	c.After(wait)
}

func goldChoices(a *model.App, s *GoldState, win *mygo.Window, c *myui.Context, now time.Time) {
	myui.Row(c).Gap(16).AlignItems(myui.Center).Children(func() {
		weight := myui.Segmented(c.Key("gold-weight"), &s.weight, gold.WeightNames()...).Label("重量")
		if weight.Changed() && a.Gold.Weight != s.weight {
			a.Gold.Weight = s.weight
			a.Gold.ApplyDefaults()
			s.weight = a.Gold.Weight
			persist(a)
			s.startSpot(win, now)
		}
		currency := myui.Select(c.Key("gold-currency"), &s.currency, model.GoldCurrencies()).Label("币种")
		if currency.Changed() && a.Gold.Currency != s.currency {
			a.Gold.Currency = s.currency
			a.Gold.ApplyDefaults()
			s.currency = a.Gold.Currency
			persist(a)
			s.startSpot(win, now)
		}
	})
}

func goldWindow(a *model.App, s *GoldState, c *myui.Context) {
	window := myui.Segmented(c.Key("gold-window"), &s.window, gold.WindowNames()...).Label("回看窗口")
	if window.Changed() && a.Gold.Window != s.window {
		a.Gold.Window = s.window
		a.Gold.ApplyDefaults()
		s.window = a.Gold.Window
		s.selected = -1
		persist(a)
	}
}

func goldQuote(s *GoldState, now time.Time, t *myui.Theme, c *myui.Context) {
	price, when, notes := spotText(s, now)
	if price != "" {
		myui.Text(c, price).FontSize(28).Bold().Selectable()
	}
	if when != "" {
		myui.Text(c, "报价时刻 "+when).TextColor(t.TextMuted).Selectable()
	}
	for _, note := range notes {
		myui.Text(c, note).TextColor(t.TextMuted)
	}
}

func spotText(s *GoldState, now time.Time) (price, when string, notes []string) {
	unit := gold.WeightUnit(s.weight)
	code, known := gold.CodeOf(s.currency)
	var amount float64
	shown := false
	if known && s.quoteOK {
		amount, shown = s.quote.Amount(code, unit)
	}
	if shown {
		price = gold.FormatQuote(amount, code, unit)
		when = gold.FormatQuoteTime(s.quote.At, time.Local)
		if gold.Expired(s.quote.At, now, s.quote.Stale, gold.SpotLimit) {
			notes = append(notes, "价格已过期")
		}
		if s.spotFailed {
			notes = append(notes, "更新失败")
		}
		return price, when, notes
	}
	if s.spotRunning || s.wantSpot {
		return "", "", []string{"正在获取金价"}
	}
	return "", "", []string{"暂时拿不到金价"}
}

func goldTrend(s *GoldState, now time.Time, t *myui.Theme, c *myui.Context) {
	unit := gold.WeightUnit(s.weight)
	dates, values := gold.Series(s.closes, now, gold.NormalizeWindow(s.window), unit)
	myui.Text(c, "美元/"+gold.WeightName(unit)).FontSize(12).TextColor(t.TextMuted)
	if note := trendNote(s, now, len(dates)); note != "" {
		myui.Text(c, note).TextColor(t.TextMuted)
	}
	if len(dates) == 0 {
		return
	}
	goldPlot(c, s, t, dates, values, unit)
}

func trendNote(s *GoldState, now time.Time, points int) string {
	if points > 0 {
		switch {
		case gold.Expired(s.histAt, now, s.histStale, gold.HistoryLimit) && s.histFailed:
			return "走势已过期 更新失败"
		case gold.Expired(s.histAt, now, s.histStale, gold.HistoryLimit):
			return "走势已过期"
		case s.histFailed:
			return "更新失败"
		default:
			return ""
		}
	}
	if s.histRunning || s.wantHist {
		return "正在获取走势"
	}
	if s.histOK {
		return "这个窗口没有日线"
	}
	return "暂时拿不到走势"
}

func goldPlot(c *myui.Context, s *GoldState, t *myui.Theme, dates []time.Time, values []float64, unit string) {
	x := chart.NewTime(dates[0], dates[len(dates)-1])
	x.Format = func(at time.Time) string {
		if at.Month() == time.January && at.Day() == 1 {
			return strconv.Itoa(at.Year())
		}
		if at.Day() == 1 {
			return strconv.Itoa(int(at.Month())) + "月"
		}
		return strconv.Itoa(int(at.Month())) + "月" + strconv.Itoa(at.Day()) + "日"
	}
	lo, hi := chart.Extent(values...)
	if lo == hi {
		pad := math.Abs(lo) * 0.05
		if pad == 0 {
			pad = 1
		}
		lo -= pad
		hi += pad
	}
	y := chart.NewLinear(lo, hi).Nice(4)
	y.Format = func(v float64) string { return gold.FormatAmount(v, unit) }
	xs := make([]float32, len(dates))
	ys := make([]float32, len(values))
	for i := range dates {
		xs[i] = x.At(dates[i])
		ys[i] = y.At(values[i])
	}
	accent := t.Accent
	border := t.Border
	label := func(_ chart.Side, tick chart.Tick) {
		myui.Text(c, tick.Label).NoWrap().FontSize(11).TextColor(t.TextMuted)
	}
	yTicks := y.Ticks(4)
	chart.Plot(c, chart.Axes{Left: yTicks, Bottom: x.Ticks(6), Label: label}, func(area myui.Element) {
		area.Draw(func(p *myui.Painter, r myui.Rect) {
			for _, tick := range yTicks {
				p.Line(r.X, chart.Y(r, tick.At), r.X+r.W, chart.Y(r, tick.At), 1, border)
			}
			p.StrokePath(chart.Line(r, xs, ys, chart.CurveLinear), 2, accent)
		})
		describe := func(i int) string {
			if i < 0 || i >= len(dates) {
				return ""
			}
			return gold.FormatReadout(dates[i], values[i], unit)
		}
		area.Label("走势图")
		chart.Select(c, area, &s.selected, xs, describe)
		if i := s.selected; i >= 0 && i < len(dates) {
			chart.RuleX(c, xs[i], 1).Background(t.TextMuted)
			chart.Point(c, xs[i], ys[i], 8, 8).Radius(4).Background(accent)
			// Grow the readout away from the nearer side so it stays on the plot.
			chart.Anchor(c, xs[i], 1, readoutAlign(xs[i]), myui.Start, func() {
				myui.Text(c, describe(i)).NoWrap().Padding(3, 8).Radius(6).Background(t.Surface)
			})
		}
	}).Height(320).MinWidth(0).Gap(6).Label("走势").Description("黄金日收盘，美元计价")
}

// readoutAlign keeps a label that hangs from the top of the plot inside it.
// Past the middle, its right edge sits on the point and the text grows left.
func readoutAlign(x float32) myui.Align {
	if x > 0.5 {
		return myui.End
	}
	return myui.Start
}

func (s *GoldState) startSpot(win *mygo.Window, now time.Time) {
	code, ok := gold.CodeOf(s.currency)
	if !ok {
		s.spotFailed = true
		return
	}
	unit := gold.WeightUnit(s.weight)
	s.spotGen++
	if s.spotCancel != nil {
		s.spotCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), gold.RequestTimeout)
	s.spotCancel = cancel
	gen := s.spotGen
	s.spotRunning = true
	s.spotFailed = false
	s.spotAt = now
	base := gold.BaseURL
	run := func() {
		quote, err := gold.FetchSpot(ctx, base, code, unit)
		publishGold(win, func() { s.finishSpot(quote, err, gen) })
	}
	if win == nil {
		run()
		return
	}
	go run()
}

func (s *GoldState) finishSpot(quote gold.Quote, err error, gen int) {
	if gen != s.spotGen || !s.onScreen {
		return
	}
	s.spotRunning = false
	if s.spotCancel != nil {
		s.spotCancel()
		s.spotCancel = nil
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		s.spotFailed = true
		return
	}
	s.quote = quote
	s.quoteOK = true
	s.spotFailed = false
}

func (s *GoldState) stopSpot() {
	s.spotGen++
	if s.spotCancel != nil {
		s.spotCancel()
		s.spotCancel = nil
	}
	s.spotRunning = false
}

func (s *GoldState) startHist(win *mygo.Window) {
	s.histGen++
	if s.histCancel != nil {
		s.histCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), gold.RequestTimeout)
	s.histCancel = cancel
	gen := s.histGen
	s.histRunning = true
	s.histFailed = false
	base := gold.BaseURL
	run := func() {
		history, err := gold.FetchHistory(ctx, base)
		publishGold(win, func() { s.finishHist(history, err, gen) })
	}
	if win == nil {
		run()
		return
	}
	go run()
}

func (s *GoldState) finishHist(history gold.History, err error, gen int) {
	if gen != s.histGen || !s.onScreen {
		return
	}
	s.histRunning = false
	if s.histCancel != nil {
		s.histCancel()
		s.histCancel = nil
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		s.histFailed = true
		return
	}
	s.closes = history.Closes
	s.histAt = history.At
	s.histStale = history.Stale
	s.histOK = true
	s.histFailed = false
	s.selected = -1
}

func (s *GoldState) stopHist() {
	s.histGen++
	if s.histCancel != nil {
		s.histCancel()
		s.histCancel = nil
	}
	s.histRunning = false
}

func publishGold(win *mygo.Window, fn func()) {
	if win == nil {
		fn()
		return
	}
	win.Update(fn)
}
