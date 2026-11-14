package app

import "testing"

func TestWindowOptions(t *testing.T) {
	opts := windowOptions(nil)
	if opts.Title != "DevBox" || opts.Width != 1200 || opts.Height != 800 || opts.MinWidth != 900 || opts.MinHeight != 600 {
		t.Fatalf("window options = %+v", opts)
	}
}
