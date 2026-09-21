package theme

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	DefaultBorder string
	Focused       string
	Navigate      string
	Insert        string
	Pending       string
	Primary       string
	PrimaryText   string
	PrimaryAlt    string
	Help          string
	HelpText      string
	Command       string
	CommandText   string
}

func Terminal() Palette {
	return Palette{
		DefaultBorder: "8",
		Focused:       "5",
		Navigate:      "2",
		Insert:        "3",
		Pending:       "3",
		Primary:       "5",
		PrimaryText:   "0",
		PrimaryAlt:    "7",
		Help:          "4",
		HelpText:      "7",
		Command:       "2",
		CommandText:   "0",
	}
}

func Color(value string) lipgloss.Color {
	return lipgloss.Color(value)
}
