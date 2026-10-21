package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/amisonnet8/crube/internal/core"
	"github.com/amisonnet8/crube/internal/search"
	"github.com/amisonnet8/crube/internal/status"
	"github.com/amisonnet8/crube/internal/textwidth"
)

// The text output (docs/reference/cli.md). Widths are display widths
// (textwidth): a wide character takes 2 columns.

const (
	gap           = 3 // spaces between the widest name and the desc
	noDescription = "! no description"
	ellipsis      = "…"
)

// row is one line of a two-column list: a name, then its desc at the column
// shared by all rows, then an optional note (tree's "(here: …)").
type row struct {
	left, desc, note string
}

// columns writes rows with their descs aligned. With width > 0 (the output is
// a terminal that wide), a desc that would make the line wider is cut and
// ends with "…"; the note is kept.
func columns(b *strings.Builder, rows []row, width int) {
	col := 0
	for _, r := range rows {
		col = max(col, textwidth.String(r.left))
	}
	col += gap
	for _, r := range rows {
		b.WriteString(r.left)
		if r.desc == "" && r.note == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(strings.Repeat(" ", col-textwidth.String(r.left)))
		desc := r.desc
		if width > 0 {
			avail := width - col
			if r.note != "" {
				avail -= 2 + textwidth.String(r.note)
			}
			desc = cut(desc, avail)
		}
		b.WriteString(desc)
		if r.note != "" {
			if desc != "" {
				b.WriteString("  ")
			}
			b.WriteString(r.note)
		}
		b.WriteString("\n")
	}
}

// cut shortens s to at most avail columns, ending with "…" when it is cut. If
// not even one character fits before "…", it is "…" alone.
func cut(s string, avail int) string {
	if textwidth.String(s) <= avail {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := textwidth.Rune(r)
		if w+rw > avail-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String() + ellipsis
}

// termWidth is the width of the terminal that Stdout is, or 0 when Stdout is
// not a terminal (a pipe or a file): then nothing is cut.
func termWidth(env Env) int {
	if env.Width != nil {
		return env.Width()
	}
	f, ok := env.Stdout.(*os.File)
	if !ok {
		return 0
	}
	fd := int(f.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}
	w, _, err := term.GetSize(fd)
	if err != nil {
		return 0
	}
	return w
}

func descOrNone(desc string) string {
	if desc == "" {
		return noDescription
	}
	return desc
}

func lsText(b *strings.Builder, res *core.LsResult, width int) {
	rows := make([]row, 0, len(res.Items))
	for _, it := range res.Items {
		rows = append(rows, row{left: it.Name, desc: descOrNone(it.Desc)})
	}
	columns(b, rows, width)
}

func treeText(b *strings.Builder, res *core.TreeResult, width int) {
	var rows []row
	var walk func(nodes []*core.TreeNode, level int)
	walk = func(nodes []*core.TreeNode, level int) {
		for _, n := range nodes {
			mark := "  "
			note := ""
			if n.Here {
				mark = "> "
				if res.Above {
					note = "(here: " + res.Here + ")"
				}
			}
			rows = append(rows, row{left: mark + strings.Repeat("  ", level) + n.Name, desc: n.Desc, note: note})
			walk(n.Children, level+1)
		}
	}
	walk(res.Entries, 0)
	columns(b, rows, width)
}

func findText(b *strings.Builder, res *search.Found, width int) {
	rows := make([]row, 0, len(res.Items))
	for _, it := range res.Items {
		rows = append(rows, row{left: it.Path, desc: it.Desc})
	}
	columns(b, rows, width)
	if res.More > 0 {
		fmt.Fprintf(b, "%s %d more (use --limit)\n", ellipsis, res.More)
	}
}

func showText(b *strings.Builder, res *core.ShowResult) {
	b.WriteString(res.Path + "\n")
	field := func(name, value string) {
		// An empty line of the detail leaves no spaces at the end.
		b.WriteString(strings.TrimRight(fmt.Sprintf("  %-8s%s", name, value), " ") + "\n")
	}
	if res.Entry != "" && res.Entry != res.Path {
		field("entry", res.Entry)
	}
	if res.Link != "" {
		field("link", res.Link)
	}
	field("desc", descOrNone(res.Desc))
	if res.Detail != "" {
		lines := strings.Split(res.Detail, "\n")
		field("detail", lines[0])
		for _, l := range lines[1:] {
			field("", l)
		}
	}
}

func statusText(b *strings.Builder, res *status.Report) {
	for _, it := range res.Items {
		switch it.Kind {
		case status.Untracked:
			fmt.Fprintf(b, "%-12s%-6s%s\n", it.Kind, it.Type, it.Path)
		case status.Invalid:
			fmt.Fprintf(b, "%-12s%s  %s %d/%d\n", it.Kind, it.Path, it.Field, it.Width, it.Limit)
		default:
			fmt.Fprintf(b, "%-12s%s\n", it.Kind, it.Path)
		}
	}
	if len(res.Items) > 0 {
		b.WriteString("\n")
	}
	fmt.Fprintf(b, "%-12sall: %s   dirs: %s\n", "coverage", level(res.Coverage.All), level(res.Coverage.Dirs))
}

func level(l status.Level) string {
	if l.OK {
		return "yes"
	}
	return fmt.Sprintf("no (%d)", l.Missing)
}
