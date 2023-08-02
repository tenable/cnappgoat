package common

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
)

var OptionsCNAPPgoatSeparateColumns = table.Options{
	DrawBorder:      false,
	SeparateColumns: true,
	SeparateFooter:  false,
	SeparateHeader:  false,
	SeparateRows:    false,
}

var ColorOptionsCNAPPgoatMagenta = table.ColorOptions{
	Footer: text.Colors{text.BgMagenta, text.FgBlack},
	Header: text.Colors{text.BgHiMagenta, text.FgBlack},
}

var TableStyleCNAPPgoatMagenta = table.Style{
	Name:    "TableStyleCNAPPgoatMagenta",
	Box:     table.StyleBoxLight,
	Color:   ColorOptionsCNAPPgoatMagenta,
	Format:  table.FormatOptionsDefault,
	Options: OptionsCNAPPgoatSeparateColumns,
	Title:   table.TitleOptionsMagentaOnBlack,
}

func GetDisplayTable() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)
	return t
}

func AppendScenarioToTable(s *cnappgoat.Scenario, t table.Writer, c int) {
	t.AppendRow(table.Row{
		c + 1,
		s.ScenarioParams.ID,
		s.ScenarioParams.FriendlyName,
		s.ScenarioParams.Platform,
		s.ScenarioParams.Module,
		s.State.String(),
	})
}
