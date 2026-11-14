package model

const (
	ThemeDark   = "dark"
	ThemeLight  = "light"
	ThemeSystem = "system"

	// NavSettings is the sidebar choice that opens the settings page.
	NavSettings = "settings"

	StatusReady = "就绪"
)

// App is the state the window rebuilds from. It holds values, never UI elements.
type App struct {
	CurrentTool string
	// Theme is dark, light, or system.
	Theme string
	// ConfigDir is MyGo's user data directory. Empty skips saving.
	ConfigDir string
	// Recent is the tool ids opened most recently, newest first.
	Recent []string
	// Favorites is the tool ids the user pinned, in the order they were added.
	Favorites []string
	// Selection is the sidebar choice: a tool id, or NavSettings.
	Selection string
	// Status, Error and Elapsed are the status bar: execution state, the
	// latest error, and how long the last operation took.
	Status  string
	Error   string
	Elapsed string
	// ShownTool is the tool the status bar last described. Switching tools
	// clears that description.
	ShownTool string
}

// New returns the initial state: JSON selected, dark theme, idle status.
func New() *App {
	a := &App{
		CurrentTool: ToolJSON,
		Selection:   ToolJSON,
		ShownTool:   ToolJSON,
		Status:      StatusReady,
	}
	a.SetTheme(ThemeDark)
	return a
}

// SetTheme stores dark, light, or system. Anything else becomes dark.
func (a *App) SetTheme(theme string) {
	switch theme {
	case ThemeLight:
		a.Theme = ThemeLight
	case ThemeSystem:
		a.Theme = ThemeSystem
	default:
		a.Theme = ThemeDark
	}
}
