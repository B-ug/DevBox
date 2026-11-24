package model

const (
	ToolJSON      = "json"
	ToolBase64    = "base64"
	ToolTimestamp = "datetime"
	ToolPDF       = "pdf"
	ToolTranslate = "translate"
	ToolTunnel    = "tunnel"
	ToolGold      = "gold"
)

// Tool is one entry in the sidebar and the workspace title that follows it.
type Tool struct {
	ID      string
	Name    string
	Summary string
}

// Tools is the navigation catalog. Recent and favorite sections list ids
// from these tools.
func Tools() []Tool {
	return []Tool{
		{ID: ToolJSON, Name: "JSON", Summary: "格式化、压缩与校验"},
		{ID: ToolBase64, Name: "Base64", Summary: "标准与 URL Safe 编解码"},
		{ID: ToolTimestamp, Name: "时间戳", Summary: "秒与毫秒转换"},
		{ID: ToolPDF, Name: "PDF", Summary: "缩小图片以减小体积"},
		{ID: ToolTranslate, Name: "翻译", Summary: "调用模型翻译文本"},
		{ID: ToolTunnel, Name: "快速隧道", Summary: "临时公开本地服务，可用邮箱限制访问"},
		{ID: ToolGold, Name: "黄金", Summary: "查看现货金价和历史走势"},
	}
}

// Lookup returns the tool with id.
func Lookup(id string) (Tool, bool) {
	for _, tool := range Tools() {
		if tool.ID == id {
			return tool, true
		}
	}
	return Tool{}, false
}
