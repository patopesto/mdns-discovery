package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
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
	BorderStyle(lg.RoundedBorder()).
	BorderForeground(lg.Color("240")).
	Foreground(lg.Color("252")).
	Align(lg.Left)

var tableHeaderStyle = Style().
	Foreground(lg.Color("203")).
	Bold(true).
	Align(lg.Left)

var tableHighlightedRowStyle = Style().
	// Bold(true).
	Background(lg.Color("96")).
	Foreground(lg.Color("255"))

// - Footer
var footerStyle = Style().
	Padding(1, 2)

// Other
var helpKeyStyle = Style().
	Foreground(lg.Color("205"))

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

// Implements help.KeyMap interface
func (m App) ShortHelp() []key.Binding {
	keys := []key.Binding{m.keys.Help}
	if m.showSettings {
		keys = append(keys, m.settings.ShortHelp()...)
	} else {
		keys = append(keys, m.table.ShortHelp()...)
		keys = append(keys, m.keys.Settings)
	}
	keys = append(keys, m.keys.Quit)
	return keys
}

// Implements help.KeyMap interface
func (m App) FullHelp() [][]key.Binding {
	var keys [][]key.Binding
	if m.showSettings {
		keys = append(keys, m.settings.FullHelp()...)
	} else {
		keys = append(keys, m.table.FullHelp()...)
		keys = append(keys, []key.Binding{m.keys.Settings})
	}
	keys = append(keys, []key.Binding{m.keys.Help, m.keys.Quit})
	return keys
}

func (m *App) viewFooter() string {
	footer := m.help.View(m)

	return footerStyle.Render(footer)
}

// Implement tea.Model interface
func (m *App) View() tea.View {
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
		m.table.SetSize(m.totalWidth, innerHeight)
		mainView = tableStyle.Render(m.table.View())
	}

	content := lg.JoinVertical(
		lg.Left,
		header,
		mainView,
		footer,
	)

	view := tea.NewView(Style().Render(content))
	view.AltScreen = true
	view.WindowTitle = APP_TITLE
	return view
}
