package tunnel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNormalizeOrigin(t *testing.T) {
	got, err := NormalizeOrigin("  localhost:8080 ")
	if err != nil || got != "http://localhost:8080" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = NormalizeOrigin("https://127.0.0.1:3000")
	if err != nil || got != "https://127.0.0.1:3000" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = NormalizeOrigin("tcp://127.0.0.1:22")
	if err != nil || got != "tcp://127.0.0.1:22" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = NormalizeOrigin("HTTP://LocalHost:9/x")
	if err != nil || got != "http://LocalHost:9/x" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = NormalizeOrigin("http://[::1]:8080")
	if err != nil || got != "http://[::1]:8080" {
		t.Fatalf("%q %v", got, err)
	}

	if _, err := NormalizeOrigin("  "); err == nil || err.Error() != errOriginRequired {
		t.Fatalf("%v", err)
	}
	if _, err := NormalizeOrigin("ftp://localhost"); err == nil || err.Error() != errOriginScheme {
		t.Fatalf("%v", err)
	}
	if _, err := NormalizeOrigin("http://"); err == nil || err.Error() != errOriginInvalid {
		t.Fatalf("%v", err)
	}
}

func TestParseEmails(t *testing.T) {
	got, err := ParseEmails(" \n")
	if err != nil || got != nil {
		t.Fatalf("%v %v", got, err)
	}
	got, err = ParseEmails(" A@Example.com \n b@example.com,*@example.com，A@example.com ")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A@Example.com", "b@example.com", "*@example.com"}
	if !slices.Equal(got, want) {
		t.Fatalf("%v", got)
	}
	if _, err := ParseEmails("not-an-email"); err == nil || err.Error() != "邮箱无效：not-an-email" {
		t.Fatalf("%v", err)
	}
	if _, err := ParseEmails("*@com"); err == nil {
		t.Fatal("single-label domain was accepted")
	}
	if _, err := ParseEmails("*@*.example.com"); err == nil {
		t.Fatal("wildcard domain was accepted")
	}

	var b strings.Builder
	for i := 0; i < 51; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "u%d@example.com", i)
	}
	if _, err := ParseEmails(b.String()); err == nil || err.Error() != errMailLimit {
		t.Fatalf("%v", err)
	}
}

func TestArgs(t *testing.T) {
	got := Args("http://localhost:8080", nil)
	want := []string{"tunnel", "--no-autoupdate", "--url", "http://localhost:8080"}
	if !slices.Equal(got, want) {
		t.Fatalf("%v", got)
	}
	got = Args("http://localhost:8080", []string{"a@b.co", "*@c.co"})
	want = append(want, "--allowed-mail", "a@b.co", "--allowed-mail", "*@c.co")
	if !slices.Equal(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestPublicURL(t *testing.T) {
	text := "INF |  https://kingston-inside-best-graphic.trycloudflare.com                                    |"
	if got := PublicURL(text); got != "https://kingston-inside-best-graphic.trycloudflare.com" {
		t.Fatalf("%q", got)
	}
	jsonLine := `{"message":"|  https://massachusetts-syndicate-effect-principal.trycloudflare.com |"}`
	if got := PublicURL(jsonLine); got != "https://massachusetts-syndicate-effect-principal.trycloudflare.com" {
		t.Fatalf("%q", got)
	}
	if PublicURL("no tunnel yet") != "" {
		t.Fatal("found a URL in unrelated text")
	}
}

func TestVersionSupportsMail(t *testing.T) {
	for _, output := range []string{
		"cloudflared version 2026.9.3 (built 2026-09-24)",
		"cloudflared version 2026.10.0",
		"cloudflared version 2027.1.0",
	} {
		if err := versionSupportsMail(output); err != nil {
			t.Fatalf("%q: %v", output, err)
		}
	}
	for _, output := range []string{
		"cloudflared version 2026.9.2",
		"cloudflared version 2025.8.1",
	} {
		err := versionSupportsMail(output)
		if err == nil || err.Error() != errVersionOld {
			t.Fatalf("%q: %v", output, err)
		}
	}
	if err := versionSupportsMail("nope"); err == nil || err.Error() != errVersionRead {
		t.Fatalf("%v", err)
	}
}

func TestResolve(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cloudflared.exe")
	if err := os.WriteFile(path, []byte("not really"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("  " + path + "  ")
	if err != nil || got != path {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := Resolve(filepath.Join(dir, "missing.exe")); err == nil || !strings.Contains(err.Error(), "找不到 cloudflared") {
		t.Fatalf("%v", err)
	}
	if _, err := Resolve(dir); err == nil {
		t.Fatal("directory was accepted")
	}
}

func TestMailSupportedCommand(t *testing.T) {
	r := Runner{NewCommand: helperCommand("version")}
	if err := r.MailSupported(context.Background(), "cloudflared"); err != nil {
		t.Fatal(err)
	}
	r.NewCommand = helperCommand("old")
	err := r.MailSupported(context.Background(), "cloudflared")
	if err == nil || err.Error() != errVersionOld {
		t.Fatalf("%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r.NewCommand = helperCommand("version")
	if err := r.MailSupported(ctx, "cloudflared"); err == nil {
		t.Fatal("canceled check ran")
	}
}

func TestStartAndStop(t *testing.T) {
	var got []string
	r := Runner{NewCommand: func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return helperCommand("url")(ctx, name, args...)
	}}
	running, err := r.Start(context.Background(), "cloudflared", "http://127.0.0.1:9", []string{"a@b.co"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(running.Stop)
	want := append([]string{"cloudflared"}, Args("http://127.0.0.1:9", []string{"a@b.co"})...)
	if !slices.Equal(got, want) {
		t.Fatalf("args %v", got)
	}
	if url := waitURL(t, running.Events); url != "https://alpha-beta-gamma-delta.trycloudflare.com" {
		t.Fatalf("url %s", url)
	}
	running.Stop()
	waitExit(t, running.Events)
}

func TestStartFailure(t *testing.T) {
	r := Runner{NewCommand: helperCommand("fail")}
	running, err := r.Start(context.Background(), "cloudflared", "http://127.0.0.1:9", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(running.Stop)
	var lines []string
	var exitErr error
	sawExit := false
	timeout := time.After(10 * time.Second)
	for !sawExit {
		select {
		case ev, ok := <-running.Events:
			if !ok {
				t.Fatal("closed before exit")
			}
			if ev.Line != "" {
				lines = append(lines, ev.Line)
			}
			if ev.Exited {
				sawExit = true
				exitErr = ev.Err
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}
	if exitErr == nil {
		t.Fatal("failure exited without an error")
	}
	if !slices.Contains(lines, "failed to request quick tunnel") {
		t.Fatalf("lines %v", lines)
	}
}

func waitURL(t *testing.T, events <-chan Event) string {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				t.Fatal("closed before the URL")
			}
			if ev.Exited {
				t.Fatalf("exited before the URL: %v", ev.Err)
			}
			if ev.URL != "" {
				return ev.URL
			}
		case <-timeout:
			t.Fatal("timeout waiting for the URL")
		}
	}
}

func waitExit(t *testing.T, events <-chan Event) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case ev, ok := <-events:
			if !ok || ev.Exited {
				return
			}
		case <-timeout:
			t.Fatal("timeout waiting for exit")
		}
	}
}

func helperCommand(mode string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHelperProcess$")
		cmd.Env = helperEnv(mode)
		return cmd
	}
}

func helperEnv(mode string) []string {
	drop := []string{"DEVBOX_TUNNEL_HELPER=", "DEVBOX_TUNNEL_MODE="}
	var env []string
	for _, item := range os.Environ() {
		skip := false
		upper := strings.ToUpper(item)
		for _, prefix := range drop {
			if strings.HasPrefix(upper, prefix) {
				skip = true
				break
			}
		}
		if !skip {
			env = append(env, item)
		}
	}
	return append(env, "DEVBOX_TUNNEL_HELPER=1", "DEVBOX_TUNNEL_MODE="+mode)
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("DEVBOX_TUNNEL_HELPER") != "1" {
		return
	}
	switch os.Getenv("DEVBOX_TUNNEL_MODE") {
	case "version":
		helperLine(os.Stdout, "cloudflared version 2026.9.3 (built 2026-09-24)")
	case "old":
		helperLine(os.Stdout, "cloudflared version 2025.8.1")
	case "url":
		helperLine(os.Stdout, "Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):")
		helperLine(os.Stdout, "|  https://alpha-beta-gamma-delta.trycloudflare.com  |")
		time.Sleep(time.Hour)
	case "fail":
		helperLine(os.Stderr, "failed to request quick tunnel")
		os.Exit(1)
	default:
		helperLine(os.Stderr, "unknown helper mode")
		os.Exit(2)
	}
	os.Exit(0)
}

func helperLine(w *os.File, text string) {
	fmt.Fprintln(w, text)
	_ = w.Sync()
}
