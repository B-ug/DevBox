package model

const (
	ThemeDark   = "dark"
	ThemeLight  = "light"
	ThemeSystem = "system"

	StatusReady    = "就绪"
	StatusSettings = "设置将在后续提供"
)

// App is the state the window rebuilds from. It holds values, never UI elements.
type App struct {
	CurrentTool string
	Theme       string
	Status      string
}

// New returns the initial state: JSON selected, dark theme, idle status.
func New() *App {
	return &App{
		CurrentTool: ToolJSON,
		Theme:       ThemeDark,
		Status:      StatusReady,
	}
}
