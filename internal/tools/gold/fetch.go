package gold

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const maxBody = 2 << 20

// BaseURL is the API origin. Tests point it at a local server.
var BaseURL = "https://xaus.com"

// ErrUnavailable means the feed has no real price or series to give.
var ErrUnavailable = errors.New("unavailable")

var client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type spotPayload struct {
	XAU struct {
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
		Unit     string  `json:"unit"`
	} `json:"xau"`
	SpotUSDOz float64 `json:"spot_usd_oz"`
	FXRate    float64 `json:"fx_rate"`
	UpdatedAt string  `json:"updated_at"`
	PriceAsOf string  `json:"price_as_of"`
	Stale     bool    `json:"stale"`
	DataState state   `json:"data_state"`
}

type historyPayload struct {
	Points []struct {
		D string  `json:"d"`
		C float64 `json:"c"`
	} `json:"points"`
	UpdatedAt string `json:"updated_at"`
	DataState state  `json:"data_state"`
}

type state struct {
	Status string `json:"status"`
	AsOf   string `json:"as_of"`
}

// FetchSpot loads one quote in currency and unit.
func FetchSpot(ctx context.Context, base, currency, unit string) (Quote, error) {
	var payload spotPayload
	if err := getJSON(ctx, spotEndpoint(base, currency, unit), &payload); err != nil {
		return Quote{}, err
	}
	return payload.quote()
}

// FetchHistory loads the daily US dollar closes, up to five years.
func FetchHistory(ctx context.Context, base string) (History, error) {
	var payload historyPayload
	if err := getJSON(ctx, strings.TrimRight(base, "/")+"/api/v1/history?days=1825", &payload); err != nil {
		return History{}, err
	}
	return payload.history()
}

func spotEndpoint(base, currency, unit string) string {
	u, err := url.Parse(strings.TrimRight(base, "/") + "/api/v1/spot")
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("currency", currency)
	q.Set("unit", unit)
	q.Set("compact", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func (p spotPayload) quote() (Quote, error) {
	if p.DataState.Status == "unavailable" || p.XAU.Price <= 0 || p.SpotUSDOz <= 0 {
		return Quote{}, ErrUnavailable
	}
	at, _ := parseTime(p.DataState.AsOf)
	if at.IsZero() {
		at, _ = parseTime(p.PriceAsOf)
	}
	if at.IsZero() {
		at, _ = parseTime(p.UpdatedAt)
	}
	fx := p.FXRate
	if fx <= 0 && strings.EqualFold(p.XAU.Currency, "USD") {
		fx = 1
	}
	return Quote{
		OK:       true,
		Price:    p.XAU.Price,
		Currency: p.XAU.Currency,
		Unit:     p.XAU.Unit,
		USDPerOz: p.SpotUSDOz,
		FX:       fx,
		At:       at,
		Stale:    p.Stale || p.DataState.Status == "stale",
	}, nil
}

func (p historyPayload) history() (History, error) {
	if p.DataState.Status == "unavailable" {
		return History{}, ErrUnavailable
	}
	closes := make([]Close, 0, len(p.Points))
	for _, point := range p.Points {
		day, err := time.Parse("2006-01-02", point.D)
		if err != nil || point.C <= 0 {
			continue
		}
		closes = append(closes, Close{
			Date:     time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC),
			USDPerOz: point.C,
		})
	}
	if len(closes) == 0 {
		return History{}, ErrUnavailable
	}
	slices.SortFunc(closes, func(a, b Close) int { return a.Date.Compare(b.Date) })
	at, _ := parseTime(p.DataState.AsOf)
	if at.IsZero() {
		at, _ = parseTime(p.UpdatedAt)
	}
	return History{
		Closes: closes,
		At:     at,
		Stale:  p.DataState.Status == "stale",
	}, nil
}

func getJSON(ctx context.Context, endpoint string, dest any) error {
	if endpoint == "" {
		return errors.New("地址无效")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.New("地址无效")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DevBox")
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		return ErrUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("请求失败")
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return err
	}
	return nil
}

func parseTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
