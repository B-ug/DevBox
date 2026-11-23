// Package tunnel starts a Cloudflare Quick Tunnel with the local cloudflared.
package tunnel

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	// maxRecipients is how many addresses or domain rules one tunnel accepts.
	maxRecipients = 50
)

const (
	errOriginRequired = "请填写要隧道的地址"
	errOriginInvalid  = "地址无效"
	errOriginScheme   = "只支持 http、https 和 tcp 地址"
	errMailInvalid    = "邮箱无效：%s"
	errMailLimit      = "邮箱最多 50 个"
	errCloudflared    = "未找到 cloudflared。请安装后加入 PATH，或选择 cloudflared.exe"
	errCloudflaredAt  = "找不到 cloudflared：%s"
	errVersionRead    = "无法识别 cloudflared 版本，邮箱限制需要 2026.9.3 或更高"
	errVersionOld     = "邮箱限制需要 cloudflared 2026.9.3 或更高版本"
)

// minMailVersion is the first cloudflared that accepts --allowed-mail.
var minMailVersion = version{2026, 9, 3}

var (
	emailPattern = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?)+$`)
	// domainPattern is *@example.com: every address at that domain.
	domainPattern    = regexp.MustCompile(`^\*@[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?)+$`)
	publicURLPattern = regexp.MustCompile(`(?i)https://[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?\.trycloudflare\.com`)
	versionPattern   = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
)

type version struct {
	year, minor, patch int
}

func (v version) atLeast(min version) bool {
	if v.year != min.year {
		return v.year > min.year
	}
	if v.minor != min.minor {
		return v.minor > min.minor
	}
	return v.patch >= min.patch
}

// NormalizeOrigin trims raw and adds http:// when the scheme is missing.
// http, https, and tcp are accepted.
func NormalizeOrigin(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New(errOriginRequired)
	}
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", errors.New(errOriginInvalid)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	switch u.Scheme {
	case "http", "https", "tcp":
	default:
		return "", errors.New(errOriginScheme)
	}
	if u.Host == "" || strings.ContainsAny(u.Host, " \t") {
		return "", errors.New(errOriginInvalid)
	}
	return u.String(), nil
}

// ParseEmails splits text on lines, commas, and spaces. A blank text allows
// everyone. *@example.com allows a whole domain. Repeats are dropped.
func ParseEmails(text string) ([]string, error) {
	fields := splitRecipients(text)
	if len(fields) == 0 {
		return nil, nil
	}
	if len(fields) > maxRecipients {
		return nil, errors.New(errMailLimit)
	}
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if !validRecipient(field) {
			return nil, fmt.Errorf(errMailInvalid, field)
		}
		key := strings.ToLower(field)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, field)
	}
	return out, nil
}

func splitRecipients(text string) []string {
	text = strings.NewReplacer("，", ",", "；", ",", ";", ",").Replace(text)
	var out []string
	for _, part := range strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ' ' || r == '\t'
	}) {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validRecipient(s string) bool {
	if strings.HasPrefix(s, "*@") {
		return domainPattern.MatchString(s)
	}
	return emailPattern.MatchString(s)
}

// Args is the cloudflared invocation for a quick tunnel. Each recipient is
// its own --allowed-mail, so the shell never has to quote them.
func Args(origin string, emails []string) []string {
	args := []string{"tunnel", "--no-autoupdate", "--url", origin}
	for _, email := range emails {
		args = append(args, "--allowed-mail", email)
	}
	return args
}

// PublicURL returns the first trycloudflare.com address in text.
func PublicURL(text string) string {
	return publicURLPattern.FindString(text)
}

// versionSupportsMail reports whether cloudflared --version output is new
// enough for --allowed-mail. An unreadable version is an error, so an old
// program is not started without the protection the user asked for.
func versionSupportsMail(output string) error {
	v, ok := parseVersion(output)
	if !ok {
		return errors.New(errVersionRead)
	}
	if !v.atLeast(minMailVersion) {
		return errors.New(errVersionOld)
	}
	return nil
}

func parseVersion(text string) (version, bool) {
	m := versionPattern.FindStringSubmatch(text)
	if m == nil {
		return version{}, false
	}
	year, err1 := strconv.Atoi(m[1])
	minor, err2 := strconv.Atoi(m[2])
	patch, err3 := strconv.Atoi(m[3])
	if err1 != nil || err2 != nil || err3 != nil {
		return version{}, false
	}
	return version{year, minor, patch}, true
}

// Resolve returns path when it names a file, or cloudflared on PATH.
func Resolve(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path != "" {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return "", fmt.Errorf(errCloudflaredAt, path)
		}
		return path, nil
	}
	found, err := exec.LookPath("cloudflared")
	if err != nil {
		return "", errors.New(errCloudflared)
	}
	return found, nil
}
