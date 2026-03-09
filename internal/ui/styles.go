package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Panel styles
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")) // Blue

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")) // Gray

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottomForeground(lipgloss.Color("240"))

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Bold(true).
			Padding(0, 1)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// State colors
	StateNewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")) // Blue

	StateActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")) // Yellow

	StateResolvedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("76")) // Green

	StateClosedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")) // Gray

	// Pipeline status colors
	SucceededStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))  // Green
	FailedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	RunningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	CanceledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray

	// Detail view
	DetailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39")).
				MarginBottom(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Width(14)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// Help
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	HelpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(1, 2)

	// Titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	// Error / Warning
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
)

// StateStyle returns the appropriate style for a work item state.
func StateStyle(state string) lipgloss.Style {
	switch state {
	case "New":
		return StateNewStyle
	case "Active":
		return StateActiveStyle
	case "Resolved":
		return StateResolvedStyle
	case "Closed", "Done", "Removed":
		return StateClosedStyle
	default:
		return NormalItemStyle
	}
}

// PipelineResultStyle returns the style for a pipeline result.
func PipelineResultStyle(result string) lipgloss.Style {
	switch result {
	case "succeeded":
		return SucceededStyle
	case "failed":
		return FailedStyle
	case "canceled":
		return CanceledStyle
	default:
		return RunningStyle
	}
}
