package tunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Event is one line of cloudflared output, or the process ending.
// URL is set when Line contains the trycloudflare.com address.
// A canceled tunnel reports Exited with a nil Err.
type Event struct {
	Line   string
	URL    string
	Exited bool
	Err    error
}

// Runner starts cloudflared. A nil NewCommand runs the real program and
// hides its console window on Windows. Tests supply NewCommand.
type Runner struct {
	NewCommand func(ctx context.Context, name string, args ...string) *exec.Cmd
}

// Running is a quick tunnel process. Stop ends it.
type Running struct {
	Events <-chan Event
	stop   func()
}

// Stop ends the process. It is safe to call more than once, and after the
// process has already exited.
func (r *Running) Stop() {
	if r == nil || r.stop == nil {
		return
	}
	r.stop()
}

// MailSupported runs cloudflared --version and accepts 2026.9.3 or newer.
// A canceled ctx returns ctx.Err() without a version error.
func (r Runner) MailSupported(ctx context.Context, bin string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cmd := r.command(ctx, bin, "--version")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if versionErr := versionSupportsMail(string(out)); versionErr != nil {
		if err != nil && strings.TrimSpace(string(out)) == "" {
			return errors.New(errVersionRead)
		}
		return versionErr
	}
	return nil
}

// Start runs cloudflared tunnel for origin. emails may be empty.
// The caller Stops the result, or cancels ctx, to end the process.
func (r Runner) Start(ctx context.Context, bin, origin string, emails []string) (*Running, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(bin) == "" {
		return nil, errors.New(errCloudflared)
	}
	ctx, cancel := context.WithCancel(ctx)
	cmd := r.command(ctx, bin, Args(origin, emails)...)
	if cmd == nil {
		cancel()
		return nil, errors.New("启动 cloudflared 失败")
	}
	if cmd.WaitDelay == 0 {
		cmd.WaitDelay = 3 * time.Second
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("启动 cloudflared 失败：%w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("启动 cloudflared 失败：%w", err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("启动 cloudflared 失败：%w", err)
	}

	events := make(chan Event, 32)
	go func() {
		defer close(events)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			readLines(ctx, stdout, events)
		}()
		go func() {
			defer wg.Done()
			readLines(ctx, stderr, events)
		}()
		wg.Wait()
		waitErr := cmd.Wait()
		ev := Event{Exited: true, Err: waitErr}
		if ctx.Err() != nil {
			ev.Err = nil
		}
		emit(ctx, events, ev)
	}()

	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		})
	}
	return &Running{Events: events, stop: stop}, nil
}

func (r Runner) command(ctx context.Context, name string, args ...string) *exec.Cmd {
	if r.NewCommand != nil {
		cmd := r.NewCommand(ctx, name, args...)
		if cmd != nil && cmd.SysProcAttr == nil {
			hideWindow(cmd)
		}
		return cmd
	}
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	return cmd
}

func readLines(ctx context.Context, r io.Reader, events chan<- Event) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		ev := Event{Line: line, URL: PublicURL(line)}
		if !emit(ctx, events, ev) {
			_, _ = io.Copy(io.Discard, r)
			return
		}
	}
}

func emit(ctx context.Context, events chan<- Event, ev Event) bool {
	select {
	case events <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}
