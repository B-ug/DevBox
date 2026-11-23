package model

import "strings"

// TunnelOrigin is the local address used when none is saved.
const TunnelOrigin = "http://localhost:8080"

// Tunnel is the quick tunnel's saved address, allowed recipients, and
// the cloudflared executable. A blank Cloudflared means PATH.
type Tunnel struct {
	Origin      string
	Emails      string
	Cloudflared string
}

// ApplyDefaults fills an empty address. Emails and Cloudflared may stay blank.
func (t *Tunnel) ApplyDefaults() {
	if strings.TrimSpace(t.Origin) == "" {
		t.Origin = TunnelOrigin
	}
}

// UseTunnel stores saved tunnel settings and fills a blank address.
func (a *App) UseTunnel(saved Tunnel) {
	a.Tunnel = saved
	a.Tunnel.ApplyDefaults()
}
