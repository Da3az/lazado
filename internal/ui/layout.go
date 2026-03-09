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

	// Frame sizes (from border style)
	HFrame int // horizontal frame (left + right border)
	VFrame int // vertical frame (top + bottom border)
}

const (
	tabBarHeight    = 2 // tab text + bottom border line
	statusBarHeight = 1
	listWidthRatio  = 0.4
)

// CalculateLayout computes the layout dimensions from the terminal size.
func CalculateLayout(width, height int) Layout {
	hFrame, vFrame := ActiveBorderStyle.GetFrameSize()

	l := Layout{
		Width:           width,
		Height:          height,
		TabBarHeight:    tabBarHeight,
		StatusBarHeight: statusBarHeight,
		HFrame:          hFrame,
		VFrame:          vFrame,
	}

	// Content area is everything between tab bar and status bar, minus border frames
	l.ContentHeight = height - tabBarHeight - statusBarHeight - vFrame
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
