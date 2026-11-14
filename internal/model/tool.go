package model

const (
	ToolJSON      = "json"
	ToolBase64    = "base64"
	ToolTimestamp = "datetime"
)

// Tool is one entry in the sidebar and the workspace title that follows it.
type Tool struct {
	ID      string
	Name    string
	Summary string
}

// Tools is the navigation catalog for this phase. Recent and favorite
// sections stay empty until those lists exist.
func Tools() []Tool {
	return []Tool{
		{ID: ToolJSON, Name: "JSON", Summary: "格式化、压缩与校验"},
		{ID: ToolBase64, Name: "Base64", Summary: "标准与 URL Safe 编解码"},
		{ID: ToolTimestamp, Name: "时间戳", Summary: "秒与毫秒转换"},
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
