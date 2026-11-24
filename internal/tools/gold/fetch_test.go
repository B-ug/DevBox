package gold

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSpotAndHistory(t *testing.T) {
	asOf := time.Date(2026, 11, 24, 1, 41, 0, 0, time.UTC).Format(time.RFC3339)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method %s", r.Method)
		}
		switch r.URL.Path {
		case "/api/v1/spot":
			q := r.URL.Query()
			if q.Get("currency") != "CNY" || q.Get("unit") != "gram" || q.Get("compact") != "1" {
				t.Errorf("query %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"xau": {"price": 788.9496, "currency": "CNY", "unit": "gram"},
				"spot_usd_oz": 3673.9,
				"fx_rate": 6.6793,
				"price_as_of": "` + asOf + `",
				"data_state": {"status": "fresh", "as_of": "` + asOf + `"}
			}`))
		case "/api/v1/history":
			if r.URL.Query().Get("days") != "1825" {
				t.Errorf("days %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{
				"points": [
					{"d": "2026-11-23", "c": 3646.8},
					{"d": "bad", "c": 1},
					{"d": "2026-11-20", "c": 3600}
				],
				"data_state": {"status": "stale", "as_of": "` + asOf + `"}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	quote, err := FetchSpot(t.Context(), server.URL, "CNY", "gram")
	if err != nil || !quote.OK || quote.Price != 788.9496 || quote.Currency != "CNY" || quote.Unit != "gram" || quote.FX != 6.6793 || quote.Stale {
		t.Fatalf("%+v err %v", quote, err)
	}
	if !quote.At.Equal(time.Date(2026, 11, 24, 1, 41, 0, 0, time.UTC)) {
		t.Fatalf("at %s", quote.At)
	}

	history, err := FetchHistory(t.Context(), server.URL)
	if err != nil || !history.Stale || len(history.Closes) != 2 || history.Closes[0].USDPerOz != 3600 || history.Closes[1].USDPerOz != 3646.8 {
		t.Fatalf("%+v err %v", history, err)
	}
}

func TestFetchUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/spot" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"points": [], "data_state": {"status": "unavailable"}}`))
	}))
	defer server.Close()
	if _, err := FetchSpot(t.Context(), server.URL, "USD", "oz"); err != ErrUnavailable {
		t.Fatalf("spot %v", err)
	}
	if _, err := FetchHistory(t.Context(), server.URL); err != ErrUnavailable {
		t.Fatalf("history %v", err)
	}
}
