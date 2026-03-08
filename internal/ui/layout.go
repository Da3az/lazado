package ui

// Layout holds the calculated dimensions for each UI region.
type Layout struct {
	Width  int
	Height int

	TabBarHeight    int
	StatusBarHeight int

	// Content area (between tab bar and status bar)
	ContentWidth  int
	ContentHeight int

	// Split between list and detail
	ListWidth   int
	DetailWidth int
}

const (
	tabBarHeight    = 1
	statusBarHeight = 1
	borderSize      = 2 // top + bottom borders
	listWidthRatio  = 0.4
)

// CalculateLayout computes the layout dimensions from the terminal size.
func CalculateLayout(width, height int) Layout {
	l := Layout{
		Width:           width,
		Height:          height,
		TabBarHeight:    tabBarHeight,
		StatusBarHeight: statusBarHeight,
	}

	// Content area is everything between tab bar and status bar
	l.ContentHeight = height - tabBarHeight - statusBarHeight - borderSize
	if l.ContentHeight < 1 {
		l.ContentHeight = 1
	}
	l.ContentWidth = width

	// Split horizontally
	l.ListWidth = int(float64(width) * listWidthRatio)
	if l.ListWidth < 20 {
		l.ListWidth = 20
	}
	l.DetailWidth = width - l.ListWidth
	if l.DetailWidth < 20 {
		l.DetailWidth = 20
	}

	return l
}
