package app

import (
	"strings"

	lg "github.com/charmbracelet/lipgloss"
)

/* ----- Styles ----- */
// Helpers
var Style = lg.NewStyle

// Window
// - Header
var headerStyle = Style().
	Padding(1, 4, 1, 2)

var titleStyle = Style().
	Bold(true).
	Foreground(lg.Color("202"))

var spinnerStyle = Style().
	PaddingRight(2).
	Foreground(lg.Color("255"))

var interfacesStyle = Style().
	Foreground(lg.Color("#606060"))

var interfaceStyle = Style().
	Padding(0, 1).
	Foreground(lg.Color("202"))

// - Table
var tableStyle = Style()

var tableBaseStyle = Style().
	BorderStyle(lg.NormalBorder()).
	BorderForeground(lg.Color("240")).
	Foreground(lg.Color("252")).
	Align(lg.Left)

var tableHeaderStyle = Style().
	Foreground(lg.Color("203")).
	Bold(true).
	Align(lg.Center)

var tableHighlightedRowStyle = Style().
	Bold(true).
	Background(lg.Color("96")).
	Foreground(lg.Color("255"))

// - Footer
var footerStyle = Style().
	Padding(1, 2)

// Other
var helpKeyStyle = Style().
	Foreground(lg.AdaptiveColor{
		Light: "#909090",
		Dark:  "205",
	})

func (m *App) viewHeader() string {
	title := titleStyle.Render(APP_TITLE)
	spinner := m.spinner.View()
	itfs := strings.Builder{}
	itfs.WriteString("interfaces ")
	for _, itf := range m.discovery.Interfaces {
		s := interfaceStyle.Render(itf.Name)
		itfs.WriteString(s)
	}
	interfaces := interfacesStyle.Render(itfs.String())

	spacerWidth := m.totalWidth - lg.Width(spinner) - lg.Width(title) - lg.Width(interfaces) - headerStyle.GetHorizontalPadding()

	header := lg.JoinHorizontal(
		lg.Center,
		spinner,
		title,
		Style().Width(spacerWidth).Render(""),
		interfaces,
	)

	return headerStyle.Render(header)
}

func (m *App) viewFooter() string {
	var footer string
	if m.showSettings {
		footer = m.help.View(m.settings)
	} else {
		footer = m.help.View(m.keys)
	}

	return footerStyle.Render(footer)
}

// Implement tea.Model interface
func (m *App) View() string {
	header := m.viewHeader()
	footer := m.viewFooter()

	// Compute height of all elements to send to component
	headerHeight := lg.Height(header)
	footerHeight := lg.Height(footer)
	innerHeight := m.totalHeight - headerHeight - footerHeight

	var mainView string
	if m.showSettings {
		m.settings.SetSize(m.totalWidth, innerHeight)
		mainView = m.settings.View()
	} else {
		pageSize := innerHeight - 6 // magic offset based on current header + tableHeader rendering
		if pageSize < 1 {
			pageSize = 1
		}
		m.table = m.table.WithMinimumHeight(innerHeight).WithPageSize(pageSize)
		mainView = tableStyle.Render(m.table.View())
	}

	view := lg.JoinVertical(
		lg.Left,
		header,
		mainView,
		footer,
	)
	return Style().Render(view)
}
