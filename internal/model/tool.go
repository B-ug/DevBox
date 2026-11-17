package model

const (
	ToolJSON      = "json"
	ToolBase64    = "base64"
	ToolTimestamp = "datetime"
	ToolPDF       = "pdf"
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
