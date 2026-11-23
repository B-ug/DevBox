package ui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"devbox/internal/model"
	"devbox/internal/tools/tunnel"

	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
)

const (
	tunnelIdle     = ""
	tunnelStarting = "starting"
	tunnelReady    = "ready"
	tunnelStopped  = "stopped"
	tunnelFailed   = "failed"

	tunnelLogLines    = 200
	tunnelVersionWait = 10 * time.Second
)

// TunnelState is the quick tunnel's process and the text the page shows.
// It lives with the window, so leaving the page does not stop the tunnel.
type TunnelState struct {
	Running   bool
	URL       string
	Log       string
	Failure   string
	Protected bool
	phase     string

	generation int
	cancel     context.CancelFunc
	proc       *tunnel.Running
}

func newTunnelState() *TunnelState {
	return &TunnelState{}
}

func tunnelWorkspace(a *model.App, editors *Editors, c *myui.Context) {
	s := editors.Tunnel
	if s == nil {
		return
	}
	t := c.Theme()
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Padding(16, 24).Gap(12).MinWidth(0).Children(func() {
				myui.Text(c, "把本地服务临时公开到 trycloudflare.com").TextColor(t.TextMuted).MinWidth(0)

				myui.Text(c, "本地地址").FontSize(13).Bold()
				origin := myui.TextInput(c.Key("tunnel-origin"), &a.Tunnel.Origin).Placeholder(model.TunnelOrigin).Label("本地地址").Grow(1).MinWidth(0).Disabled(s.Running)
				if origin.Changed() && !s.Running {
					persist(a)
				}

				myui.Text(c, "允许访问的邮箱").FontSize(13).Bold()
				emails := myui.TextArea(c.Key("tunnel-emails"), &a.Tunnel.Emails).Placeholder("留空则任何人都能打开链接").Label("允许访问的邮箱").MinHeight(96).MaxHeight(160).MinWidth(0).Disabled(s.Running)
				if emails.Changed() && !s.Running {
					persist(a)
				}
				myui.Text(c, "每行一个邮箱，或 *@example.com。留空则不限制。邮箱限制需要 cloudflared 2026.9.3 或更高版本。").FontSize(12).TextColor(t.TextMuted).MinWidth(0)

				myui.Text(c, "cloudflared 路径").FontSize(13).Bold()
				myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
					path := myui.TextInput(c.Key("tunnel-path"), &a.Tunnel.Cloudflared).Placeholder("留空则从 PATH 查找").Label("cloudflared 路径").Grow(1).MinWidth(0).Disabled(s.Running)
					if path.Changed() && !s.Running {
						persist(a)
					}
					if myui.Button(c.Key("tunnel-browse"), "浏览").Disabled(s.Running).Shrink(0).Clicked() {
						if picked, ok := pickCloudflared(a, a.Tunnel.Cloudflared); ok {
							a.Tunnel.Cloudflared = picked
							persist(a)
						}
					}
				})

				switch s.phase {
				case tunnelStarting:
					myui.Text(c, "正在建立隧道").TextColor(t.TextMuted)
				case tunnelReady:
					myui.Text(c, "公网地址").FontSize(13).Bold()
					myui.Text(c, s.URL).MinWidth(0)
					access := "任何人打开链接都能访问"
					if s.Protected {
						access = "访客需要通过邮箱验证码"
					}
					myui.Text(c, access).TextColor(t.TextMuted).MinWidth(0)
				case tunnelStopped:
					myui.Text(c, "已停止").TextColor(t.TextMuted)
				case tunnelFailed:
					if s.Failure != "" {
						myui.Text(c, s.Failure).TextColor(t.Danger).MinWidth(0)
					}
				}

				myui.Text(c, "日志").FontSize(13).Bold()
				myui.TextArea(c.Key("tunnel-log"), &s.Log).ReadOnly(true).Label("日志").Placeholder("cloudflared 的输出会显示在这里").MinHeight(160).MinWidth(0)
			})
		})
		myui.Divider(c)
		myui.Column(c).Shrink(0).Padding(12, 24).Gap(8).Children(func() {
			myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
				if myui.PrimaryButton(c.Key("tunnel-start"), "启动").Disabled(s.Running).Shrink(0).Clicked() {
					startTunnel(a, s, editors.window)
				}
				if myui.Button(c.Key("tunnel-stop"), "停止").Disabled(!s.Running).Shrink(0).Clicked() {
					s.Stop(a)
				}
				if myui.Button(c.Key("tunnel-copy"), "复制链接").Disabled(s.URL == "").Shrink(0).Clicked() {
					copyTunnel(a, s, c)
				}
				if myui.Button(c.Key("tunnel-open"), "在浏览器打开").Disabled(s.URL == "").Shrink(0).Clicked() {
					openTunnel(a, s, c)
				}
			})
			myui.Text(c, "点「停止」或退出程序后，公网链接失效。修改地址或邮箱需要先停止，再重新启动。").FontSize(12).TextColor(t.TextMuted).MinWidth(0)
		})
	})
}

func copyTunnel(a *model.App, s *TunnelState, c *myui.Context) {
	if s == nil || s.URL == "" {
		return
	}
	c.WriteClipboard(s.URL)
	s.setStatus(a, "已复制", "")
}

func openTunnel(a *model.App, s *TunnelState, c *myui.Context) {
	if s == nil || s.URL == "" {
		return
	}
	c.OpenURL(s.URL)
	s.setStatus(a, "已打开", "")
}

func pickCloudflared(a *model.App, current string) (string, bool) {
	if !mygo.App.IsReady() {
		return "", false
	}
	dir := current
	if current != "" {
		dir = filepath.Dir(current)
	}
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:      mygo.FocusedWindow(),
		Title:       "选择 cloudflared",
		DefaultPath: dir,
		Filters:     []mygo.FileFilter{{Name: "cloudflared", Extensions: []string{"exe"}}},
	})
	if err != nil {
		a.Status = "失败"
		a.Error = err.Error()
		a.Elapsed = ""
		return "", false
	}
	if len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}

func startTunnel(a *model.App, s *TunnelState, win *mygo.Window) {
	if a == nil || s == nil || s.Running {
		return
	}
	origin, err := tunnel.NormalizeOrigin(a.Tunnel.Origin)
	if err != nil {
		failTunnel(a, err)
		return
	}
	emails, err := tunnel.ParseEmails(a.Tunnel.Emails)
	if err != nil {
		failTunnel(a, err)
		return
	}
	bin, err := tunnel.Resolve(a.Tunnel.Cloudflared)
	if err != nil {
		failTunnel(a, err)
		return
	}
	a.Tunnel.Origin = origin
	persist(a)

	if s.cancel != nil {
		s.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.generation++
	gen := s.generation
	s.cancel = cancel
	s.proc = nil
	s.Running = true
	s.URL = ""
	s.Failure = ""
	s.Protected = len(emails) > 0
	s.phase = tunnelStarting
	s.Log = ""
	s.setStatus(a, "正在建立隧道", "")

	run := func() {
		runner := tunnel.Runner{}
		if len(emails) > 0 {
			vctx, vcancel := context.WithTimeout(ctx, tunnelVersionWait)
			err := runner.MailSupported(vctx, bin)
			vcancel()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if errors.Is(err, context.DeadlineExceeded) {
					err = errors.New("无法读取 cloudflared 版本")
				}
				publishTunnel(win, func() { s.fail(a, gen, err) })
				return
			}
		}
		if ctx.Err() != nil {
			return
		}
		running, err := runner.Start(ctx, bin, origin, emails)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			publishTunnel(win, func() { s.fail(a, gen, err) })
			return
		}
		publishTunnel(win, func() {
			if s.generation != gen || !s.Running {
				running.Stop()
				return
			}
			s.proc = running
		})
		for ev := range running.Events {
			ev := ev
			publishTunnel(win, func() { s.apply(a, gen, ev) })
		}
	}
	go run()
}

func failTunnel(a *model.App, err error) {
	if a == nil || err == nil {
		return
	}
	a.Status = "失败"
	a.Error = err.Error()
	a.Elapsed = ""
}

// Stop ends cloudflared and clears the public address. A nil app skips the
// status bar, which is what closing the window uses.
func (s *TunnelState) Stop(a *model.App) {
	if s == nil || !s.Running {
		return
	}
	s.generation++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.proc != nil {
		s.proc.Stop()
		s.proc = nil
	}
	s.Running = false
	s.URL = ""
	s.Failure = ""
	s.Protected = false
	s.phase = tunnelStopped
	s.Log = appendLog(s.Log, "已停止")
	s.setStatus(a, "已停止", "")
}

func (s *TunnelState) fail(a *model.App, gen int, err error) {
	if s.generation != gen || err == nil {
		return
	}
	s.Running = false
	s.URL = ""
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.proc != nil {
		s.proc.Stop()
		s.proc = nil
	}
	s.Failure = err.Error()
	s.phase = tunnelFailed
	s.setStatus(a, "失败", err.Error())
}

func (s *TunnelState) apply(a *model.App, gen int, ev tunnel.Event) {
	if s.generation != gen {
		return
	}
	if ev.Line != "" {
		s.Log = appendLog(s.Log, ev.Line)
	}
	if ev.URL != "" && s.URL == "" {
		s.URL = ev.URL
		s.Failure = ""
		s.phase = tunnelReady
		s.setStatus(a, "隧道已就绪", "")
	}
	if !ev.Exited {
		return
	}
	s.Running = false
	s.proc = nil
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.URL = ""
	msg := "隧道已结束"
	if ev.Err != nil {
		msg = failureText(s.Log)
	}
	s.Failure = msg
	s.phase = tunnelFailed
	s.setStatus(a, "失败", msg)
}

func (s *TunnelState) setStatus(a *model.App, status, errText string) {
	if a == nil || a.CurrentTool != model.ToolTunnel {
		return
	}
	a.Status = status
	a.Error = errText
	a.Elapsed = ""
}

func publishTunnel(win *mygo.Window, fn func()) {
	if win == nil {
		fn()
		return
	}
	win.Update(fn)
}

func appendLog(log, line string) string {
	line = strings.TrimRight(line, "\r")
	if strings.TrimSpace(line) == "" {
		return log
	}
	if log == "" {
		log = line
	} else {
		log += "\n" + line
	}
	if strings.Count(log, "\n")+1 <= tunnelLogLines {
		return log
	}
	parts := strings.Split(log, "\n")
	if len(parts) > tunnelLogLines {
		parts = parts[len(parts)-tunnelLogLines:]
	}
	return strings.Join(parts, "\n")
}

func failureText(log string) string {
	lines := strings.Split(log, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.Contains(line, "trycloudflare.com") {
			continue
		}
		if strings.Trim(line, "|+- ") == "" {
			continue
		}
		return line
	}
	return "隧道已结束"
}
