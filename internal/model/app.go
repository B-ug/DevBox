package model

const (
	ThemeDark   = "dark"
	ThemeLight  = "light"
	ThemeSystem = "system"

	ThemeDarkIndex  = 0
	ThemeLightIndex = 1

	StatusReady = "就绪"
)

// App is the state the window rebuilds from. It holds values, never UI elements.
type App struct {
	CurrentTool string
	// Theme is dark or light. ThemeIndex is what the toolbar segmented
	// control binds, and SetTheme keeps the two together.
	Theme      string
	ThemeIndex int
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
		ShownTool:   ToolJSON,
		Status:      StatusReady,
	}
	a.SetTheme(ThemeDark)
	return a
}

// SetTheme stores a dark or light appearance. Anything else becomes dark.
// Following the system theme waits for the settings page.
func (a *App) SetTheme(theme string) {
	switch theme {
	case ThemeLight:
		a.Theme = ThemeLight
		a.ThemeIndex = ThemeLightIndex
	default:
		a.Theme = ThemeDark
		a.ThemeIndex = ThemeDarkIndex
	}
}
