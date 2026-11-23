package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func TestTunnelPage(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("快速隧道"); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"临时公开本地服务，可用邮箱限制访问",
		"本地地址",
		"允许访问的邮箱",
		"cloudflared 路径",
		"启动",
		"停止",
		"复制链接",
		"在浏览器打开",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if a.Tunnel.Origin != model.TunnelOrigin {
		t.Fatalf("origin %q", a.Tunnel.Origin)
	}
	if buttonDisabled(t, tt, "停止") == false {
		t.Fatal("stop was enabled before the tunnel started")
	}
	if buttonDisabled(t, tt, "启动") {
		t.Fatal("start was disabled")
	}
	if buttonDisabled(t, tt, "复制链接") == false || buttonDisabled(t, tt, "在浏览器打开") == false {
		t.Fatal("link actions were enabled without a URL")
	}

	a.Tunnel.Emails = "not-an-email"
	tt.Frame()
	if err := tt.Click("启动"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "失败" || a.Error != "邮箱无效：not-an-email" || editors.Tunnel.Running {
		t.Fatalf("status %q error %q running %v", a.Status, a.Error, editors.Tunnel.Running)
	}

	a.Tunnel.Emails = ""
	a.Tunnel.Origin = "ftp://localhost"
	tt.Frame()
	if err := tt.Click("启动"); err != nil {
		t.Fatal(err)
	}
	if a.Error != "只支持 http、https 和 tcp 地址" {
		t.Fatalf("error %q", a.Error)
	}

	a.Tunnel.Origin = model.TunnelOrigin
	a.Tunnel.Cloudflared = filepath.Join(t.TempDir(), "missing.exe")
	tt.Frame()
	if err := tt.Click("启动"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.Error, "找不到 cloudflared") || editors.Tunnel.Running {
		t.Fatalf("status %q error %q running %v", a.Status, a.Error, editors.Tunnel.Running)
	}
}

func TestTunnelStopButton(t *testing.T) {
	a := model.New()
	a.CurrentTool = model.ToolTunnel
	a.Selection = model.ToolTunnel
	a.ShownTool = model.ToolTunnel
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("快速隧道"); err != nil {
		t.Fatal(err)
	}
	editors.Tunnel.Running = true
	editors.Tunnel.URL = "https://alpha-beta-gamma-delta.trycloudflare.com"
	editors.Tunnel.Protected = true
	editors.Tunnel.phase = tunnelReady
	tt.Frame()
	if !tt.HasText(editors.Tunnel.URL) || !tt.HasText("访客需要通过邮箱验证码") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if buttonDisabled(t, tt, "停止") || buttonDisabled(t, tt, "启动") == false {
		t.Fatal("buttons did not follow the running tunnel")
	}
	if buttonDisabled(t, tt, "复制链接") || buttonDisabled(t, tt, "在浏览器打开") {
		t.Fatal("link actions stayed disabled")
	}

	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	if !editors.Tunnel.Running || editors.Tunnel.URL == "" {
		t.Fatal("switching tools stopped the tunnel")
	}
	if err := tt.Click("快速隧道"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("https://alpha-beta-gamma-delta.trycloudflare.com") {
		t.Fatalf("url lost after returning: %q", tt.Texts())
	}

	if err := tt.Click("停止"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if editors.Tunnel.Running || editors.Tunnel.URL != "" || editors.Tunnel.phase != tunnelStopped {
		t.Fatalf("%+v", editors.Tunnel)
	}
	if a.Status != "已停止" || !tt.HasText("已停止") {
		t.Fatalf("status %q texts %q", a.Status, tt.Texts())
	}
	if buttonDisabled(t, tt, "停止") == false {
		t.Fatal("stop stayed enabled")
	}
}

func TestEditorsCloseStopsTunnel(t *testing.T) {
	editors := NewEditors()
	editors.Tunnel.Running = true
	editors.Tunnel.URL = "https://alpha-beta-gamma-delta.trycloudflare.com"
	editors.Tunnel.phase = tunnelReady
	editors.Close()
	if editors.Tunnel.Running || editors.Tunnel.URL != "" || editors.Tunnel.phase != tunnelStopped {
		t.Fatalf("%+v", editors.Tunnel)
	}
}

func TestAppendLogCapsLines(t *testing.T) {
	var log string
	for i := 0; i < tunnelLogLines+10; i++ {
		log = appendLog(log, "line")
	}
	if strings.Count(log, "\n")+1 != tunnelLogLines {
		t.Fatalf("lines %d", strings.Count(log, "\n")+1)
	}
	if got := failureText("|\nhttps://a.trycloudflare.com\nfailed to request quick tunnel"); got != "failed to request quick tunnel" {
		t.Fatalf("%q", got)
	}
	if got := failureText(""); got != "隧道已结束" {
		t.Fatalf("%q", got)
	}
}

func buttonDisabled(t *testing.T, tt *myui.Tester, name string) bool {
	t.Helper()
	info, ok := tt.AccessByRole(myui.RoleButton, name)
	if !ok {
		t.Fatalf("no button %s in %q", name, tt.Texts())
	}
	return info.Disabled
}
