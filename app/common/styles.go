package common

import (
	"image/color"

	lg "charm.land/lipgloss/v2"
)

type Styles struct {
	// Colors
	Color struct {
		Top    color.Color
		Mid    color.Color
		MidLow color.Color
		Bottom color.Color

		Text      color.Color
		Highlight color.Color
		Lowlight  color.Color

		Grey25 color.Color
		Grey50 color.Color
		Grey75 color.Color
	}

	// Styles
	Base lg.Style

	Header struct {
		Base       lg.Style
		Title      lg.Style
		Spinner    lg.Style
		Interfaces lg.Style
		Interface  lg.Style
	}

	Table struct {
		Base     lg.Style
		Header   lg.Style
		Selected lg.Style
	}

	Settings struct {
		Base lg.Style
	}

	Footer struct {
		Base lg.Style
		Help lg.Style
	}
}

func NewStyles() (s Styles) {
	// Colors
	s.Color.Top = lg.Color("202")
	s.Color.Mid = lg.Color("203")
	s.Color.MidLow = lg.Color("204")
	s.Color.Bottom = lg.Color("205")

	s.Color.Text = lg.Color("252")
	s.Color.Highlight = lg.Color("204")
	s.Color.Lowlight = lg.Color("96")

	s.Color.Grey25 = lg.Color("250")
	s.Color.Grey50 = lg.Color("244")
	s.Color.Grey75 = lg.Color("239")

	// Styles
	s.Base = lg.NewStyle()

	s.Header.Base = lg.NewStyle().
		Padding(1, 4, 1, 2)

	s.Header.Title = lg.NewStyle().
		Foreground(s.Color.Top).
		Bold(true)

	s.Header.Spinner = lg.NewStyle().
		PaddingRight(2).
		Foreground(lg.Color("255"))

	s.Header.Interfaces = lg.NewStyle().
		Foreground(s.Color.Grey50)

	s.Header.Interface = lg.NewStyle().
		Padding(0, 1).
		Foreground(s.Color.Top)

	s.Table.Base = lg.NewStyle().
		BorderStyle(lg.RoundedBorder()).
		BorderForeground(s.Color.Grey75).
		Foreground(s.Color.Text).
		Align(lg.Left)

	s.Table.Header = lg.NewStyle().
		Border(lg.RoundedBorder(), false, false, true, false).
		BorderForeground(s.Color.Grey75).
		Foreground(s.Color.Mid).
		Bold(true).
		Align(lg.Left)

	s.Table.Selected = lg.NewStyle().
		Background(s.Color.Lowlight).
		Foreground(lg.Color("255"))
		// Bold(true)

	s.Settings.Base = lg.NewStyle()

	s.Footer.Base = lg.NewStyle().
		Padding(1, 2)

	s.Footer.Help = lg.NewStyle().
		Foreground(s.Color.Bottom)

	return s
}

var DefaultStyles Styles

func init() {
	DefaultStyles = NewStyles()
}
